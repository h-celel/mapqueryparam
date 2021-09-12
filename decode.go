package mapqueryparam

import (
	"encoding/json"
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
		return newDecodeError("must decode to pointer", "", nil)
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()

		if val.IsNil() {
			val.Set(reflect.New(t))
		}

		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		return newDecodeError(fmt.Sprintf("cannot decode into value of type: %s", t.String()), "", nil)
	}

	newVal := reflect.New(t)

	err := decodeFields(query, newVal.Elem())
	if err != nil {
		return err
	}

	val.Set(newVal.Elem())

	return nil
}

// decodeFields iterates over the fields of the value passed to it, decodes the values appropriate for the field, and
// stores the values in the field.
func decodeFields(query map[string][]string, val reflect.Value) error {
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		f := t.Field(i)

		// don't decode to unexported fields
		isUnexported := f.PkgPath != ""
		if isUnexported {
			continue
		}

		fTyp := f.Type
		fVal := val.Field(i)

		// iterate over embedded fields
		if f.Anonymous {
			for fTyp.Kind() == reflect.Ptr {
				fTyp = fTyp.Elem()

				if fVal.IsNil() {
					fVal.Set(reflect.New(fTyp))
				}

				fVal = fVal.Elem()
			}

			err := decodeFields(query, fVal)
			if err != nil {
				return err
			}
		}

		var s []string
		var tag string
		var ok bool

		fieldTags := getFieldTags(f)
		for _, tag = range fieldTags {
			if s, ok = query[tag]; ok {
				break
			}
		}
		if len(s) == 0 {
			continue
		}

		err := decodeField(s, fVal)
		if err != nil {
			return newDecodeError(fmt.Sprintf("unable to decode value in field '%s'", tag), tag, err)
		}
	}
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
