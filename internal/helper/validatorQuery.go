package helper

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func ExtractQueryTags[T any](v T) map[string]int {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("query")
		if tag != "" {
			fields[tag] = i
		}
	}
	return fields
}

// Populate struct from query params
func PopulateFromQuery[T any](query url.Values) (T, error) {
	var data T
	val := reflect.ValueOf(&data).Elem()
	allowedKeys := ExtractQueryTags(data)

	for key, values := range query {
		fieldIndex, ok := allowedKeys[key]
		if !ok {
			return data, fmt.Errorf("query parameter '%s' not allowed", key)
		}

		field := val.Field(fieldIndex)
		if !field.CanSet() {
			continue
		}

		// collect values: support both ?a=1,2 and ?a=1&a=2
		var allVals []string
		for _, v := range values {
			allVals = append(allVals, strings.Split(v, ",")...)
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(allVals[0])

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.Atoi(allVals[0])
			if err != nil {
				return data, fmt.Errorf("invalid int for %s: %v", key, err)
			}
			field.SetInt(int64(intVal))

		case reflect.Bool:
			boolVal, err := strconv.ParseBool(allVals[0])
			if err != nil {
				return data, fmt.Errorf("invalid bool for %s: %v", key, err)
			}
			field.SetBool(boolVal)

		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				t, err := ParseTime(allVals[0])
				if err != nil {
					return data, fmt.Errorf("invalid time for %s: %v", key, err)
				}
				field.Set(reflect.ValueOf(t))
			} else {
				return data, fmt.Errorf("unsupported struct type: %s", field.Type())
			}

		case reflect.Slice:
			elemKind := field.Type().Elem().Kind()
			switch elemKind {
			case reflect.String:
				field.Set(reflect.ValueOf(allVals))
			case reflect.Int:
				var intSlice []int
				for _, r := range allVals {
					iv, err := strconv.Atoi(r)
					if err != nil {
						return data, fmt.Errorf("invalid int in slice for %s: %v", key, err)
					}
					intSlice = append(intSlice, iv)
				}
				field.Set(reflect.ValueOf(intSlice))
			default:
				return data, fmt.Errorf("unsupported slice type: %s", elemKind)
			}

		default:
			return data, fmt.Errorf("unsupported type: %s", field.Kind())
		}
	}

	return data, nil
}

func ValidateQuery[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	queryParams, _ := url.ParseQuery(r.URL.RawQuery)
	data, err := PopulateFromQuery[T](queryParams)

	if err != nil {
		ErrorWriter(w, r, http.StatusBadRequest, err)
		return nil, err
	}

	return &data, nil

}
