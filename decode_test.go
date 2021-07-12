package mapqueryparam_test

import (
	"mapqueryparam"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	type args struct {
		query map[string][]string
		v     interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			"Empty", args{map[string][]string{},
				func() *struct{} {
					s := struct{}{}
					return &s
				}()},
			func() *struct{} {
				s := struct{}{}
				return &s
			}(), false,
		},

		{"FailAddressable", args{map[string][]string{}, struct{}{}}, struct{}{}, true},

		{
			"InitializeNil", args{map[string][]string{},
				func() *struct{} {
					var s struct{}
					return &s
				}()},
			func() *struct{} {
				s := struct{}{}
				return &s
			}(), false,
		},

		{
			"BasicStruct", args{map[string][]string{"Value": {"foobar"}},
				func() *struct{ Value string } {
					s := struct{ Value string }{}
					return &s
				}()},
			func() *struct{ Value string } {
				s := struct{ Value string }{"foobar"}
				return &s
			}(), false,
		},

		{
			"Strings", args{map[string][]string{"A": {"foo"}, "B": {"bar"}},
				func() *struct{ A, B string } {
					s := struct{ A, B string }{}
					return &s
				}()},
			func() *struct{ A, B string } {
				s := struct{ A, B string }{"foo", "bar"}
				return &s
			}(), false,
		},

		{
			"Integers", args{map[string][]string{"A": {"2"}, "B": {"3"}},
				func() *struct{ A, B int } {
					s := struct{ A, B int }{}
					return &s
				}()},
			func() *struct{ A, B int } {
				s := struct{ A, B int }{2, 3}
				return &s
			}(), false,
		},

		{
			"Uintegers", args{map[string][]string{"A": {"2"}, "B": {"3"}},
				func() *struct{ A, B uint } {
					s := struct{ A, B uint }{}
					return &s
				}()},
			func() *struct{ A, B uint } {
				s := struct{ A, B uint }{2, 3}
				return &s
			}(), false,
		},

		{
			"Floats", args{map[string][]string{"A": {"2.2"}, "B": {"3.3"}},
				func() *struct{ A, B float64 } {
					s := struct{ A, B float64 }{}
					return &s
				}()},
			func() *struct{ A, B float64 } {
				s := struct{ A, B float64 }{2.2, 3.3}
				return &s
			}(), false,
		},

		{
			"Maps", args{map[string][]string{"A": {"{\"foo\":\"bar\"}"}},
				func() *struct{ A map[string]string } {
					s := struct{ A map[string]string }{}
					return &s
				}()},
			func() *struct{ A map[string]string } {
				s := struct{ A map[string]string }{map[string]string{"foo": "bar"}}
				return &s
			}(), false,
		},

		{
			"Structs", args{map[string][]string{"A": {"{\"B\":\"foobar\"}"}},
				func() *struct{ A struct{ B string } } {
					s := struct{ A struct{ B string } }{}
					return &s
				}()},
			func() *struct{ A struct{ B string } } {
				s := struct{ A struct{ B string } }{struct{ B string }{"foobar"}}
				return &s
			}(), false,
		},

		{
			"Arrays", args{map[string][]string{"A": {"foo", "bar"}},
				func() *struct{ A [2]string } {
					s := struct{ A [2]string }{}
					return &s
				}()},
			func() *struct{ A [2]string } {
				s := struct{ A [2]string }{[2]string{"foo", "bar"}}
				return &s
			}(), false,
		},

		{
			"Slices", args{map[string][]string{"A": {"foo", "bar"}},
				func() *struct{ A []string } {
					s := struct{ A []string }{}
					return &s
				}()},
			func() *struct{ A []string } {
				s := struct{ A []string }{[]string{"foo", "bar"}}
				return &s
			}(), false,
		},

		{
			"Interface", args{map[string][]string{"B": {"foo", "bar"}},
				func() *interface{ A() []string } {
					var i interface{ A() []string }
					return &i
				}()},
			func() *interface{ A() []string } {
				var i interface{ A() []string }
				return &i
			}(), true,
		},

		// reflect.DeepEqual does not like time.Time :(
		/*
			{
				"Times", args{map[string][]string{"A": {time.Unix(1000, 1000).Format(time.RFC3339Nano)}, "B": {"1000"}, "C": {string(mustMarshal(time.Unix(1000, 1000).MarshalJSON()))}},
					func() *struct{ A, B, C time.Time } {
						s := struct{ A, B, C time.Time }{}
						return &s
					}()},
				func() *struct{ A, B, C time.Time } {
					s := struct{ A, B, C time.Time }{time.Unix(1000, 1000), time.Unix(1000, 0), time.Unix(1000, 1000)}
					return &s
				}(), false,
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mapqueryparam.Decode(tt.args.query, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.v, tt.want) {
				t.Errorf("Encode() got = %v, want %v", tt.args.v, tt.want)
			}
		})
	}
}

func mustMarshal(b []byte, _ error) []byte {
	return b
}
