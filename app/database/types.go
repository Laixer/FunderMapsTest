package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type StringArray []string

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	switch v := value.(type) {
	// case []byte:
	// 	str := string(v)
	// 	str = strings.Trim(str, "{}")
	// 	if str == "" {
	// 		*a = []string{}
	// 	} else {
	// 		*a = strings.Split(str, ",")
	// 	}
	// 	return nil
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

type JSONObject map[string]interface{}

func (j *JSONObject) Scan(value interface{}) error {
	if value == nil {
		*j = JSONObject{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	var result map[string]interface{}
	err := json.Unmarshal(bytes, &result)
	*j = JSONObject(result)
	return err
}

func (j JSONObject) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

type JSONArray []interface{}

func (a *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*a = []interface{}{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	var result []interface{}
	err := json.Unmarshal(bytes, &result)
	*a = JSONArray(result)
	return err
}

func (a JSONArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(a)
	return string(valueString), err
}
