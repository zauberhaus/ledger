package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/creasty/defaults"
	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/ec-systems/core.ledger.tool/pkg/logger"
	"github.com/ec-systems/core.ledger.tool/pkg/types"

	"github.com/fsnotify/fsnotify"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type RootCommand struct {
	cobra.Command
	cfgFile string
	version *Version
}

func GetRootCmd(version *Version) *RootCommand {
	var rootCmd *RootCommand

	rootCmd = &RootCommand{
		Command: cobra.Command{Use: "core.ledger.tool",
			Short:         "EC core workflow worker",
			SilenceErrors: true,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				err := rootCmd.initializeConfig(cmd)
				if err != nil {
					return err
				}

				cfg := config.Configuration()
				logger.SetLogLevel(cfg.LogLevel)

				verbose, err := cmd.Flags().GetBool("verbose")
				if err == nil && verbose {
					logger.Info(cfg.String())
				}

				if cfg.Assets != nil && len(cfg.Assets) > 0 {

					_ = cfg.Assets
				}

				return nil
			}},
		version: version,
	}

	err := rootCmd.init()
	if err != nil {
		logger.Fatal(err)
	}

	addVersionCmd(rootCmd)
	addAssetsCmd(rootCmd)
	addAddCmd(rootCmd)
	addRemoveCmd(rootCmd)
	addTxCmd(rootCmd)
	addAccountsCmd(rootCmd)
	addCustomerCmd(rootCmd)
	addKeysCmd(rootCmd)
	addHistoryCmd(rootCmd)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (r *RootCommand) Execute() error {
	return r.Command.Execute()
}

func (r *RootCommand) init() error {
	cfg := config.Configuration()

	// generate env variables based on tags
	AutoBindEnv(cfg)

	// set default values based on tags
	err := defaults.Set(cfg)
	if err != nil {
		return err
	}

	r.PersistentFlags().StringVar(&r.cfgFile, "config", "", "Config file (default is $HOME/"+r.Use+".yaml)")

	r.PersistentFlags().StringP("log", "l", cfg.LogLevel.String(), "Log level ("+strings.Join(cfg.LogLevel.Names(), ", ")+")")
	r.bind("LogLevel", "log")

	r.PersistentFlags().BoolP("verbose", "v", false, "Verbose logging")

	r.PersistentFlags().StringP("user", "u", cfg.ClientOptions.Username, "Database user")
	r.bind("ClientOptions.Username", "user")

	r.PersistentFlags().StringP("password", "p", cfg.ClientOptions.Password, "Database user password")
	r.bind("ClientOptions.Password", "password")

	r.PersistentFlags().StringP("database", "d", cfg.ClientOptions.Database, "Database name")
	r.bind("ClientOptions.Database", "database")

	r.PersistentFlags().BoolP("mtls", "m", cfg.ClientOptions.MTLs, "Enable mtls")
	r.bind("ClientOptions.MTLs", "mtls")

	r.PersistentFlags().String("certificate", cfg.ClientOptions.MTLsOptions.Certificate, "MTLs certificate file name")
	r.bind("ClientOptions.MTLsOptions.Certificate", "certificate")

	r.PersistentFlags().String("pkey", cfg.ClientOptions.MTLsOptions.Pkey, "MTLs key file name")
	r.bind("ClientOptions.MTLsOptions.Pkey", "pkey")

	r.PersistentFlags().String("ca", cfg.ClientOptions.MTLsOptions.ClientCAs, "MTLs ca file name")
	r.bind("ClientOptions.MTLsOptions.ClientCAs", "ca")

	r.PersistentFlags().StringToString("assets", types.DefaultAssetMap.Map(), "Supported assets")
	r.bind("Assets", "assets")

	//r.PersistentFlags().StringToInt("statuses", types.DefaultStatusMap.Map(), "Supported statuses")

	statuses := types.DefaultStatusMap

	r.PersistentFlags().Var(&statuses, "statuses", "Supported statuses")
	r.bind("Statuses", "statuses")

	return nil
}

func (r *RootCommand) initializeConfig(cmd *cobra.Command) error {
	if r.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(r.cfgFile)
	} else {
		tmp := os.Getenv("CONFIG")
		if tmp != "" {
			// Use config file from env variables.
			r.cfgFile = tmp
			viper.SetConfigFile(r.cfgFile)
		} else {

			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				logger.Error("Get homedir: %v", err)
				os.Exit(1)
			}

			viper.AddConfigPath(home)
			viper.SetConfigName(r.Use)
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info(fmt.Sprintf("Using config file: %v", viper.ConfigFileUsed()))

		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			logger.Info(fmt.Sprintf("Config file changed: %v", e.Name))
		})
	} else {
		// file not found isn't an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	options := mapstructure.ComposeDecodeHookFunc(
		config.DurationHookFunc(),
		logger.LogLevelHookFunc(),
		types.StatusHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	err := viper.Unmarshal(config.Configuration(), viper.DecodeHook(options))
	if err != nil {
		return fmt.Errorf("Unmarshal config file: %v", err)
	}

	return nil
}

func (r *RootCommand) GetVersion() *Version {
	return r.version
}

func (r *RootCommand) bind(target string, source string) {
	flag := r.PersistentFlags().Lookup(source)
	if flag == nil {
		logger.Error("Flag not found", target, source)
		return
	}

	viper.BindPFlag(target, flag)
}

func (r *RootCommand) bindCmdFlag(flags *pflag.FlagSet, target string, source string) {
	flag := flags.Lookup(source)
	if flag == nil {
		logger.Error("Flag not found", target, source)
		return
	}

	viper.BindPFlag(target, flag)
}

func (r *RootCommand) EnvBindings() map[string][]string {
	v := viper.GetViper()
	f := reflect.ValueOf(v).Elem().FieldByName("env")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	i := rf.Interface()
	return i.(map[string][]string)
}
