package mapqueryparam

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Encode(v interface{}) (map[string][]string, error) {
	if v == nil {
		return map[string][]string{}, nil
	}

	res := make(map[string][]string)

	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return map[string][]string{}, nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("unable to encode non-struct")
	}
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fTyp := t.Field(i)
		isUnexported := fTyp.PkgPath != ""
		if isUnexported {
			continue
		}

		fVal := val.Field(i)

		if isEmptyValue(fVal) {
			continue
		}

		d, err := encodeField(fVal)
		if err != nil {
			return nil, err
		}
		if len(d) == 0 {
			continue
		}

		res[fTyp.Name] = d
	}

	return res, nil
}

func encodeField(v reflect.Value) ([]string, error) {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		res := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			s, err := encodeValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			res[i] = s
		}
		return res, nil
	case reflect.Interface, reflect.Ptr:
		return encodeField(v.Elem())
	default:
		s, err := encodeValue(v)
		if err != nil {
			return nil, err
		}
		return []string{s}, nil
	}
}

func encodeValue(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Map, reflect.Struct:
		b, err := json.Marshal(v.Interface())
		return string(b), err
	case reflect.Interface, reflect.Ptr:
		return encodeValue(v.Elem())
	default:
		return "", fmt.Errorf("unsupported field kind: %d", v.Kind())
	}
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
