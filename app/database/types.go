package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type StringArray []string

func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		str := string(v)
		str = strings.Trim(str, "{}")
		if str == "" {
			*a = []string{}
		} else {
			*a = strings.Split(str, ",")
		}
		return nil
	case string:
		str := strings.Trim(v, "{}")
		if str == "" {
			*a = []string{}
		} else {
			*a = strings.Split(str, ",")
		}
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type StringArray", value)
	}
}

func (a StringArray) Value() (driver.Value, error) {
	var cleanedArray []string
	for _, str := range a {
		if strings.TrimSpace(str) != "" {
			cleanedArray = append(cleanedArray, str)
		}
	}
	if len(cleanedArray) == 0 {
		return nil, nil
	}
	return fmt.Sprintf("{%s}", strings.Join(cleanedArray, ",")), nil
}

type JSONObject map[string]any

func (j *JSONObject) Scan(value any) error {
	if value == nil {
		*j = JSONObject{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type JSONObject", value)
	}

	var result map[string]any
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSONObject: %w", err)
	}
	*j = JSONObject(result)
	return nil
}

func (j JSONObject) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	valueBytes, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONObject: %w", err)
	}
	return string(valueBytes), nil
}

type JSONArray []any

func (a *JSONArray) Scan(value any) error {
	if value == nil {
		*a = []any{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type JSONArray", value)
	}

	var result []any
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSONArray: %w", err)
	}
	*a = JSONArray(result)
	return nil
}

func (a JSONArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	valueBytes, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONArray: %w", err)
	}
	return string(valueBytes), nil
}
