package mapqueryparam

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// EncodeValues takes a input struct and encodes the content into the form of a set of query parameters.
// Input must be a pointer to a struct. Same as Encode.
func EncodeValues(v interface{}) (url.Values, error) {
	return Encode(v)
}

// Encode takes a input struct and encodes the content into the form of a set of query parameters.
// Input must be a pointer to a struct. Same as EncodeValues.
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
	err := encodeFields(val, res)
	if err != nil {
		return res, err
	}

	return res, nil
}

// encodeFields iterates over the fields of the value passed to it, and stores the encoded fields in the results map.
func encodeFields(val reflect.Value, result map[string][]string) error {
	for i := 0; i < val.NumField(); i++ {
		fTyp := val.Type().Field(i)

		// don't encode unexported fields
		isUnexported := fTyp.PkgPath != ""
		if isUnexported {
			continue
		}

		// don't attempt to encode empty fields
		fVal := val.Field(i)
		if isEmptyValue(fVal) {
			continue
		}

		// iterate over embedded fields
		if fTyp.Anonymous {
			for fVal.Kind() == reflect.Ptr {
				if fVal.IsNil() {
					continue
				}
				fVal = fVal.Elem()
			}
			err := encodeFields(fVal, result)
			if err != nil {
				return err
			}
			continue
		}

		d, err := encodeField(fVal)
		if err != nil {
			return err
		}
		if len(d) == 0 {
			continue
		}

		fieldTags := getFieldTags(fTyp)

		result[fieldTags[0]] = d
	}
	return nil
}

// getFieldTags returns the tags or names that a struct field is identified by. It prioritizes the MQP tag over the
// json tag. It defaults to the field name if neither tag is available.
func getFieldTags(t reflect.StructField) (res []string) {
	if tags := t.Tag.Get(mapQueryParameterTagName); len(tags) > 0 {
		for _, s := range strings.Split(tags, ",") {
			if len(s) > 0 {
				res = append(res, s)
			}
		}
	}

	// ignore json tags and field name if mqp tag is present
	if len(res) > 0 {
		return
	}

	if tags := t.Tag.Get("json"); len(tags) > 0 {
		jsonTags := strings.Split(tags, ",")
		if len(jsonTags) > 0 && len(jsonTags[0]) > 0 {
			res = append(res, jsonTags[0])
		}
	}

	// ignore field name if json tag is present
	if len(res) > 0 {
		return
	}

	res = append(res, t.Name)

	return
}

// encodeField encodes a field of the input struct as a set of parameter strings. Arrays and slices are represented as
// multiple strings. Other values are encoded as a single string
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

// encodeValue encodes a single value as a string. Base types are formatted using `strconv`. Maps and structs are
// encoded as json objects using standard json marshaling. Channels and functions are skipped, as they're not supported.
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
	case reflect.Complex64:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 64), nil
	case reflect.Complex128:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 128), nil
	case reflect.Map, reflect.Struct:
		i := v.Interface()
		switch t := i.(type) {
		case time.Time:
			return t.Format(time.RFC3339Nano), nil
		default:
			b, err := json.Marshal(i)
			return string(b), err
		}
	case reflect.Interface, reflect.Ptr:
		return encodeValue(v.Elem())
	case reflect.Chan, reflect.Func:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported field kind: %s", v.Kind().String())
	}
}

// isEmptyValue validated whether a value is empty/zero/nil. Used to determine if a field should be omitted from the
// encoded result.
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
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Chan, reflect.Func:
		return true
	case reflect.Struct:
		i := v.Interface()
		switch t := i.(type) {
		case time.Time:
			return t.IsZero()
		}
	}
	return false
}
