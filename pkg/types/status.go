//go:generate go run github.com/ec-systems/core.ledger.tool/pkg/generator/statuses/

package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

var (
	statuses Statuses = DefaultStatusMap

	AllStatuses Status = -1
)

type Statuses map[string]Status

func (s Statuses) Parse(txt string) (Status, error) {
	status, ok := s[txt]
	if !ok {
		return Status(-1), fmt.Errorf("unsupported status: %v", txt)
	}

	return status, nil
}

func (s Statuses) Map() map[string]int {
	m := map[string]int{}

	for k, v := range s {
		m[k] = int(v)
	}

	return m
}

func (s Statuses) String() string {
	data, err := s.MarshalText2()
	if err != nil {
		return ""
	}

	return string(data)
}

func (s *Statuses) Set(text string) error {
	return s.UnmarshalText2([]byte(text))
}

func (s Statuses) Type() string {
	return "Statuses"
}

func (s Statuses) MarshalText2() ([]byte, error) {
	m := map[string]int{}

	for k, v := range s {
		m[k] = int(v)
	}

	data, err := json.Marshal(&m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Statuses) UnmarshalText2(data []byte) error {
	m := map[string]int{}

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	statuses := map[string]Status{}
	for k, v := range m {
		statuses[k] = Status(v)
	}

	*s = statuses

	return nil
}

func StatusHookFunc() mapstructure.DecodeHookFuncType {
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
		if t != reflect.TypeOf(DefaultStatusMap) {
			return data, nil
		}

		statusses := Statuses{}

		err := statusses.Set(data.(string))
		if err != nil {
			return nil, err
		}

		// Format/decode/parse the data and return the new value
		return statusses, nil
	}
}

type Status int

func (s Status) String() string {
	if s == -1 {
		return ""
	}

	for k, v := range statuses {
		if v == s {
			return k
		}
	}

	return "Unknown status"
}

func (s Status) Valid() bool {
	for _, v := range statuses {
		if s == v {
			return true
		}
	}

	return false
}

func (s Status) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%v", s)), nil
}

func (s *Status) UnmarshalText(text []byte) error {
	if id, err := strconv.Atoi(string(text)); err == nil {
		status := Status(id)
		if !status.Valid() {
			return fmt.Errorf("invalid status: %v", id)
		}
		*s = status
	} else {
		tmp, err := statuses.Parse(string(text))
		if err != nil {
			return err
		}
		*s = tmp
	}

	return nil
}
