package types

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
)

type Format uint16

const (
	JSON      Format = 1
	Protobuf  Format = 2
	TSVPacked Format = 3
)

var Formats = map[string]Format{
	"json":     JSON,
	"protobuf": Protobuf,
}

func (s Format) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%v", uint16(s))), nil
}

func (s *Format) GetFlag() pflag.Value {
	return s
}

func (s *Format) UnmarshalText(text []byte) error {
	if id, err := strconv.Atoi(string(text)); err == nil {
		format := Format(id)
		if !format.Valid() {
			return fmt.Errorf("unknown format: %v", id)
		}
		*s = format
	} else {
		tmp, ok := Formats[string(text)]
		if !ok {
			return fmt.Errorf("unknown format: %v", text)
		}
		*s = tmp
	}

	return nil
}

func (s Format) Valid() bool {
	for _, v := range Formats {
		if v == s {
			return true
		}
	}

	return false
}

func (s Format) String() string {
	for k, v := range Formats {
		if v == s {
			return k
		}
	}

	return "unknown"
}

func (s *Format) Set(text string) error {
	format, ok := Formats[text]
	if !ok {
		return fmt.Errorf("unknown database value format: %v", text)
	}

	*s = format

	return nil
}

func (s Format) Type() string {
	return "Format"
}

func FormatHookFunc() mapstructure.DecodeHookFuncType {
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
		if t != reflect.TypeOf(Format(0)) {
			return data, nil
		}

		format := Protobuf

		err := format.Set(data.(string))
		if err != nil {
			return nil, err
		}

		// Format/decode/parse the data and return the new value
		return format, nil
	}
}
