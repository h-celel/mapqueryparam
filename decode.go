package mapqueryparam

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Decode(query map[string][]string, v interface{}) error {
	val := reflect.ValueOf(v)
	t := reflect.TypeOf(v)

	elemVal := val
	for elemVal.Kind() == reflect.Ptr {
		elemVal = elemVal.Elem()
	}

	elemT := t
	for elemT.Kind() == reflect.Ptr {
		elemT = elemT.Elem()
	}

	if !elemVal.CanAddr() {
		return errors.New("unable to address return value")
	}

	for i := 0; i < elemVal.NumField(); i++ {
		fTyp := elemT.Field(i)
		isUnexported := fTyp.PkgPath != ""
		if isUnexported {
			continue
		}

		var s []string
		var ok bool
		if s, ok = query[fTyp.Name]; !ok {
			continue
		}

		fVal := elemVal.Field(i)
		err := decodeField(s, fVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func decodeField(s []string, v reflect.Value) error {
	if len(s) < 0 {
		return nil
	}
	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len() && i < len(s); i++ {
			iVal := v.Index(i)
			err := decodeValue(s[i], iVal.Addr())
			if err != nil {
				return err
			}
		}
	case reflect.Slice:
		sVal := reflect.New(v.Type()).Elem()
		for i := 0; i < len(s); i++ {
			iVal := reflect.New(v.Type().Elem())
			err := decodeValue(s[i], iVal)
			if err != nil {
				return err
			}
			sVal = reflect.Append(sVal, iVal.Elem())
		}
		v.Set(sVal)
	default:
		return decodeValue(s[0], v.Addr())
	}
	return nil
}

func decodeValue(s string, v reflect.Value) error {
	switch v.Elem().Kind() {
	case reflect.String:
		v.Elem().SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.Elem().SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.Elem().SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.Elem().SetFloat(f)
	case reflect.Map, reflect.Struct:
		err := json.Unmarshal([]byte(s), v.Interface())
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported field kind: %s", v.Kind().String())
	}
	return nil
}
