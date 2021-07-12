package mapqueryparam

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

func DecodeValues(query url.Values, v interface{}) error {
	return Decode(query, v)
}

func Decode(query map[string][]string, v interface{}) error {
	val := reflect.ValueOf(v)
	t := reflect.TypeOf(v)

	if val.Kind() != reflect.Ptr {
		return errors.New("must decode to pointer")
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()

		if val.IsNil() {
			val.Set(reflect.New(t))
		}

		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("cannot decode into value of type: %s", t.String())
	}

	newVal := reflect.New(t)

	for i := 0; i < newVal.Elem().NumField(); i++ {
		fTyp := t.Field(i)
		isUnexported := fTyp.PkgPath != ""
		if isUnexported {
			continue
		}

		var s []string
		var ok bool
		if s, ok = query[fTyp.Name]; !ok {
			continue
		}

		fVal := newVal.Elem().Field(i)
		err := decodeField(s, fVal)
		if err != nil {
			return err
		}
	}

	val.Set(newVal.Elem())

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
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.Elem().SetBool(b)
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
		return fmt.Errorf("unsupported field kind: %s", v.Elem().Kind().String())
	}
	return nil
}
