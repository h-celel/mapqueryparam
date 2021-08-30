package mapqueryparam_test

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/h-celel/mapqueryparam"
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
			"NilPointers", args{map[string][]string{"Value": {"foobar"}},
				func() *struct{ Value *string } {
					var s struct{ Value *string }
					return &s
				}()},
			func() *struct{ Value *string } {
				st := "foobar"
				s := struct{ Value *string }{&st}
				return &s
			}(), false,
		},

		{
			"PointerToPointer", args{map[string][]string{"Value": {"foobar"}},
				func() *struct{ Value **string } {
					var s struct{ Value **string }
					return &s
				}()},
			func() *struct{ Value **string } {
				st := "foobar"
				ptr := &st
				s := struct{ Value **string }{&ptr}
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
			"ComplexNums", args{map[string][]string{"A": {"(2+2i)"}, "B": {"(3+3i)"}},
				func() *struct{ A, B complex128 } {
					s := struct{ A, B complex128 }{}
					return &s
				}()},
			func() *struct{ A, B complex128 } {
				s := struct{ A, B complex128 }{2 + 2i, 3 + 3i}
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

		{
			"JsonTag", args{map[string][]string{"b": {"foobar"}},
				func() *struct {
					A string `json:"b,omitempty"`
				} {
					s := struct {
						A string `json:"b,omitempty"`
					}{}
					return &s
				}()},
			func() *struct {
				A string `json:"b,omitempty"`
			} {
				s := struct {
					A string `json:"b,omitempty"`
				}{"foobar"}
				return &s
			}(), false,
		},

		{
			"MQPTag", args{map[string][]string{"b": {"foobar"}},
				func() *struct {
					A string `mqp:"b"`
				} {
					s := struct {
						A string `mqp:"b"`
					}{}
					return &s
				}()},
			func() *struct {
				A string `mqp:"b"`
			} {
				s := struct {
					A string `mqp:"b"`
				}{"foobar"}
				return &s
			}(), false,
		},

		{
			"MultipleMQPTag", args{map[string][]string{"c": {"foobar"}},
				func() *struct {
					A string `mqp:"b,c"`
				} {
					s := struct {
						A string `mqp:"b,c"`
					}{}
					return &s
				}()},
			func() *struct {
				A string `mqp:"b,c"`
			} {
				s := struct {
					A string `mqp:"b,c"`
				}{"foobar"}
				return &s
			}(), false,
		},

		{
			"SkipChannels", args{map[string][]string{"Value": {"foobar"}},
				func() *struct{ Value chan string } {
					s := struct{ Value chan string }{}
					return &s
				}()},
			func() *struct{ Value chan string } {
				s := struct{ Value chan string }{}
				return &s
			}(), false,
		},

		{
			"SkipFunctions", args{map[string][]string{"Value": {"foobar"}},
				func() *struct{ Value func() } {
					s := struct{ Value func() }{}
					return &s
				}()},
			func() *struct{ Value func() } {
				s := struct{ Value func() }{}
				return &s
			}(), false,
		},
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

func TestDecodeTime(t *testing.T) {
	// Test Decode() against time.Parse() of the values, using RFC3339Nano
	tests := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05+07:00",
		"2006-01-02T15:04:05-07:00",
	}

	for _, test := range tests {
		values := url.Values{"time": []string{test}}

		actual := struct {
			Time time.Time `json:"time"`
		}{}

		if err := mapqueryparam.DecodeValues(values, &actual); err != nil {
			t.Fatalf("decode failed: %s", err)
		}

		expected, err := time.Parse(time.RFC3339Nano, test)
		if err != nil {
			t.Fatalf("failed to parse expected time: %e", err)
		}

		if !expected.Equal(actual.Time) {
			t.Errorf("testing %q: wanted %q, got %q", test, expected, actual.Time)
		}
	}
}
