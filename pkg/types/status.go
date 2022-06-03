//go:generate go run github.com/ec-systems/core.ledger.service/pkg/generator/statuses/

package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

var (
	AllStatuses Status = -1
)

type Statuses map[string]Status

func (s Statuses) Parse(text string) (Status, error) {
	if id, err := strconv.Atoi(string(text)); err == nil {
		status := Status(id)
		if !s.IsValid(status) {
			return 0, fmt.Errorf("invalid status: %v", id)
		}
		return status, nil
	} else {
		status, ok := s[string(text)]
		if !ok {
			return 0, fmt.Errorf("status %v not found", text)
		}
		return status, nil
	}
}

func (s Statuses) IsValid(status Status) bool {
	for _, v := range s {
		if v == status {
			return true
		}
	}

	return false
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

func (s Statuses) MarshalJSON() ([]byte, error) {
	output := map[string]int64{}

	for name, status := range s {
		output[name] = int64(status)
	}
	return json.Marshal(output)
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

func (s Status) String(st Statuses) string {
	if s == -1 {
		return "Unknown"
	}

	for k, v := range st {
		if v == s {
			return k
		}
	}

	return "Unknown status"
}
