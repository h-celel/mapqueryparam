# mapqueryparam 

[![go report card](https://goreportcard.com/badge/github.com/h-celel/mapqueryparam "go report card")](https://goreportcard.com/report/github.com/h-celel/mapqueryparam)
[![test status](https://github.com/h-celel/mapqueryparam/actions/workflows/tests.yml/badge.svg "tests")](https://github.com/h-celel/mapqueryparam/actions/workflows/tests.yml)
[![golangci-lint](https://github.com/h-celel/mapqueryparam/actions/workflows/golangci-lint.yml/badge.svg "golangci-lint")](https://github.com/h-celel/mapqueryparam/actions/workflows/golangci-lint.yml)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go)](https://pkg.go.dev/github.com/h-celel/mapqueryparam)

mapqueryparam is a Go library for encoding and decoding structs as URL query
parameters. The query parameters use the same format as the ones found in the 
`net/url` library. All basic values, arrays and slices are encoded as single 
or multiple string values. Maps and structs are encoded using json encoding.

This encoding omits empty/zero/nil values in all cases, as there is no 
convention for representing the difference between them in the standard 
query parameter format.

Channels and function types cannot be encoded. 

Cyclic data structures will cause the encoder to get stuck in an infinite loop.

## Installation

```
$ go get github.com/h-celel/mapqueryparam
```


## Usage

Below are some examples how the library is used. Check
[example.go](https://github.com/h-celel/mapqueryparam/blob/master/example/example.go)
for more examples.


### Encode

```go
o := StructOfChoice{}

parameters, err := mapqueryparam.EncodeValues(&o)
if err != nil {
    panic(err)
}

url := fmt.Sprintf("some.site?%s", parameters.Encode())
```


### Decode

```go
var o StructOfChoice

err := mapqueryparam.DecodeValues(req.Query(), &o)
if err != nil {
    panic(err)
}
```
