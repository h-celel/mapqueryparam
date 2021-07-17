# mapqueryparam

mapqueryparam is a Go library for encoding and decoding structs as URL query
parameters. The query parameters use the same format as the ones found in the 
`net/url` library. All basic values, arrays and slices are encoded as single 
or multiple string values. Maps and structs are encoded using json encoding.

This encoding omits empty/zero/nil values in all cases, as there is no 
convention for representing the difference between them in the standard 
query parameter format.

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

req.Query()

err := mapqueryparam.DecodeValues(req.Query(), &o)
if err != nil {
    panic(err)
}
```