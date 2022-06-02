//go:generate go run github.com/ec-systems/core.ledger.service/pkg/generator/statuses/

package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
	list := []string{}

	for k, v := range s {
		list = append(list, fmt.Sprintf("%v=%v", k, int(v)))
	}

	return strings.Join(list, ",")
}

func (s *Statuses) Set(text string) error {
	pairs := strings.Split(text, ",")
	statuses := map[string]Status{}

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			key := strings.Trim(kv[0], " ")
			id, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return err
			}

			statuses[key] = Status(id)
		}
	}

	*s = statuses

	return nil
}

func (s Statuses) Type() string {
	return "Statuses"
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

		statuses := Statuses{}

		err := statuses.Set(data.(string))
		if err != nil {
			return nil, err
		}

		// Format/decode/parse the data and return the new value
		return statuses, nil
	}
}

type Status int64

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
	return []byte(fmt.Sprintf("%v", int(s))), nil
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
