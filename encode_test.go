package mapqueryparam_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/h-celel/mapqueryparam"
)

func TestEncode(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{"EmptyInput", args{struct{}{}}, map[string][]string{}, false},
		{"NilInput", args{nil}, map[string][]string{}, false},
		{"NonStruct", args{"foobar"}, nil, true},
		{"BasicStruct", args{struct{ Value string }{"foobar"}}, map[string][]string{"Value": {"foobar"}}, false},
		{"PointerToStruct", args{func() interface{} {
			s := struct{ Value string }{"foobar"}
			return &s
		}()}, map[string][]string{"Value": {"foobar"}}, false},
		{"PointerToPointer", args{func() interface{} {
			s := struct{ Value string }{"foobar"}
			p := &s
			return &p
		}()}, map[string][]string{"Value": {"foobar"}}, false},
		{"SkipUnexportedFields", args{struct{ value string }{"foobar"}}, map[string][]string{}, false},
		{"SkipZeroFields", args{struct{ Value string }{""}}, map[string][]string{}, false},
		{"Arrays", args{struct{ Value [2]string }{[2]string{"foo", "bar"}}}, map[string][]string{"Value": {"foo", "bar"}}, false},
		{"Slices", args{struct{ Value []string }{[]string{"foo", "bar"}}}, map[string][]string{"Value": {"foo", "bar"}}, false},
		{"Integers", args{struct{ Value int32 }{32}}, map[string][]string{"Value": {"32"}}, false},
		{"UIntegers", args{struct{ Value uint32 }{32}}, map[string][]string{"Value": {"32"}}, false},
		{"Floats", args{struct{ Value float32 }{32.32}}, map[string][]string{"Value": {"32.32"}}, false},
		{"ComplexNums", args{struct{ Value complex128 }{1.0 + 2.0i}}, map[string][]string{"Value": {"(1+2i)"}}, false},
		{"Bool", args{struct{ Value bool }{true}}, map[string][]string{"Value": {"true"}}, false},
		{"Structs", args{struct{ Value struct{ Value2 string } }{struct{ Value2 string }{"foobar"}}}, map[string][]string{"Value": {"{\"Value2\":\"foobar\"}"}}, false},
		{"Maps", args{struct{ Value map[string]string }{map[string]string{"Value2": "foobar"}}}, map[string][]string{"Value": {"{\"Value2\":\"foobar\"}"}}, false},
		{"Times", args{struct{ Value time.Time }{time.Unix(1000, 1000)}}, map[string][]string{"Value": {time.Unix(1000, 1000).Format(time.RFC3339Nano)}}, false},
		{"JsonTag", args{struct {
			A string `json:"b,omitempty"`
		}{"foobar"}}, map[string][]string{"b": {"foobar"}}, false},
		{"MQPTag", args{struct {
			A string `json:"b" mqp:"c,d"`
		}{"foobar"}}, map[string][]string{"c": {"foobar"}}, false},
		{"SkipChannels", args{struct{ Value chan string }{Value: make(chan string)}}, map[string][]string{}, false},
		{"SkipFunctions", args{struct{ Value func() }{Value: func() {}}}, map[string][]string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapqueryparam.Encode(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
