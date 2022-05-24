package cmd

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/spf13/viper"
)

func AutoBindEnv(config interface{}) {
	parseTags(viper.GetViper(), reflect.ValueOf(config).Type(), []string{}, []string{})
}

func parseTags(viper *viper.Viper, fieldType reflect.Type, path []string, envpath []string) {
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	for i := 0; i < fieldType.NumField(); i++ {
		f := fieldType.Field(i)
		tag := f.Tag.Get("env")
		t := f.Type

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		switch t.Kind() {
		case reflect.Struct:
			subEnvPath := envpath
			if tag != "" {
				if tag != "-" {
					subEnvPath = append(envpath, strings.ToUpper(tag))
				}
			} else if !f.Anonymous {
				subEnvPath = append(envpath, getEnvName(f.Name))
			}

			subPath := append(path, f.Name)
			parseTags(viper, t, subPath, subEnvPath)
		default:
			tmp := append(path, f.Name)
			name := strings.Join(tmp, ".")

			if tag != "" {
				if tag != "-" {
					tmp = append(envpath, strings.ToUpper(tag))
					tag := strings.Join(tmp, "_")

					viper.BindEnv(name, tag)
				}
			} else {
				tmp = append(envpath, getEnvName(f.Name))
				envVar := strings.Join(tmp, "_")

				viper.BindEnv(name, envVar)
			}
		}
	}
}

func getEnvName(name string) string {
	letters := []rune{rune(name[0])}
	lastWasUpper := true

	for _, c := range name[1:] {
		if unicode.IsUpper(c) {
			if !lastWasUpper {
				letters = append(letters, '_', c)
				lastWasUpper = true
			} else {
				letters = append(letters, c)
			}
		} else if c == '.' {
			letters = append(letters, '_')
			lastWasUpper = false
		} else {
			letters = append(letters, c)
			lastWasUpper = false
		}
	}

	result := strings.ToUpper(string(letters))

	return result
}
