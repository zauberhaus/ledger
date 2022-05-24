package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration().String(), nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	err := unmarshal(&value)
	if err != nil {
		return err
	}

	tmp, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	*d = Duration(tmp)

	return nil
}

func (d Duration) MarshalTOML() ([]byte, error) {
	return []byte("\"" + time.Duration(d).String() + "\""), nil
}

func (d *Duration) UnmarshalTOML(node interface{}) error {
	value, ok := node.(string)
	if !ok {
		return fmt.Errorf("Invalid duration string: %v", node)
	}

	tmp, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	*d = Duration(tmp)

	return nil
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func DurationHookFunc() mapstructure.DecodeHookFuncType {
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
		if t != reflect.TypeOf(Duration(time.Minute)) {
			return data, nil
		}

		val, err := time.ParseDuration(data.(string))
		if err != nil {
			return nil, err
		}

		// Format/decode/parse the data and return the new value
		return Duration(val), nil
	}
}
