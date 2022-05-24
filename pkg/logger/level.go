package logger

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Level int

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel

	PanicLevel Level = iota + 4
	FatalLevel
)

func ParseLevel(text string) (Level, error) {
	var level Level
	err := level.UnmarshalText([]byte(text))
	return level, err
}

func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return nil
	}

	switch strings.ToLower(string(text)) {
	case "debug":
		*l = DebugLevel
	case "info", "":
		*l = InfoLevel
	case "warn":
		*l = WarnLevel
	case "error":
		*l = ErrorLevel
	case "panic":
		*l = PanicLevel
	case "fatal":
		*l = FatalLevel
	default:
		return fmt.Errorf("Unknown log level: %v", string(text))
	}

	return nil
}

func (l Level) String() string {
	text, ok := l.All()[l]
	if ok {
		return text
	} else {
		return fmt.Sprintf("Level(%d)", l)
	}
}

func (l Level) All() map[Level]string {
	return map[Level]string{
		DebugLevel: "debug",
		InfoLevel:  "info",
		WarnLevel:  "warn",
		ErrorLevel: "error",
		PanicLevel: "panic",
		FatalLevel: "fatal",
	}
}

func (l Level) Names() []string {
	names := []string{}
	all := l.All()

	for _, v := range all {
		names = append(names, v)
	}

	return names
}

func LogLevelHookFunc() mapstructure.DecodeHookFuncType {
	// Wrapped in a function call to add optional input parameters (eg. separator)
	return func(
		f reflect.Type, // data type
		t reflect.Type, // target data type
		data interface{}, // raw data
	) (interface{}, error) {
		// Check if the data type matches the expected one
		if f.Kind() != reflect.String {
			return data, nil
		}

		// Check if the target type matches the expected one
		if t != reflect.TypeOf(InfoLevel) {
			return data, nil
		}

		val, err := ParseLevel(data.(string))
		if err != nil {
			return nil, err
		}

		// Format/decode/parse the data and return the new value
		return val, nil
	}
}
