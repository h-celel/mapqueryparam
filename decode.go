package mapqueryparam

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// DecodeValues takes a set of query parameters and uses reflection to decode the content into an output structure.
// Output must be a pointer to a struct. Same as Decode.
func DecodeValues(query url.Values, v interface{}) error {
	return Decode(query, v)
}

// Decode takes a set of query parameters and uses reflection to decode the content into an output structure.
// Output must be a pointer to a struct. Same as DecodeValues.
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

		fieldTag := getFieldTag(fTyp)
		if s, ok = query[fieldTag]; !ok {
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

// decodeField decodes a set of parameter strings as a field of the output struct. Arrays and slices are represented as
// multiple values. Other values are decoded as a single value.
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
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decodeField(s, v.Elem())
	default:
		return decodeValue(s[0], v.Addr())
	}
	return nil
}

// decodeValue decodes a parameter string as a value. Base types are parsed using `strconv`. Maps and structs are
// decoded as json objects using standard json unmarshaling. Channels and functions are skipped, as they're not
// supported.
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
	case reflect.Complex64, reflect.Complex128:
		f, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return err
		}
		v.Elem().SetComplex(f)
	case reflect.Map, reflect.Struct:
		i := v.Interface()
		switch i.(type) {
		case *time.Time:
			t, err := parseTime(s)
			if err != nil {
				return err
			}
			v.Elem().Set(reflect.ValueOf(t))
		default:
			err := json.Unmarshal([]byte(s), i)
			if err != nil {
				return err
			}
		}
	case reflect.Chan, reflect.Func:
	default:
		return fmt.Errorf("unsupported field kind: %s", v.Elem().Kind().String())
	}
	return nil
}

// parseTime parses a string as time.Time. It supports the RFC3339 format, unix seconds, and json marshalled time.Time
// structs.
func parseTime(s string) (time.Time, error) {
	// attempt to parse time as RFC3339 string
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}

	// attempt to parse time as float number of unix seconds
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		sec, dec := math.Modf(f)
		return time.Unix(int64(sec), int64(dec*(1e9))), nil
	}

	// attempt to parse time as json marshaled value
	if err := json.Unmarshal([]byte(s), &t); err == nil {
		return t, nil
	}

	return time.Time{}, err
}
