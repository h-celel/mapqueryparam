package main

import (
	"fmt"
	"log"

	"github.com/h-celel/mapqueryparam"
)

func main() {
	type CoolStruct struct {
		Name     string
		ID       int
		ParentID *int
	}

	type CoolRoot struct {
		Names     [2]string
		secret    string
		ID        int
		CoolParts []CoolStruct
		Aliases   []string
		IsCool    bool
	}

	cr := &CoolRoot{
		Names:  [2]string{"Mr", "Cool"},
		secret: "hush",
		ID:     32,
		CoolParts: []CoolStruct{
			{
				Name: "Very cool",
				ID:   12,
			},
			{
				Name: "Not so cool",
				ID:   45,
			},
		},
		Aliases: []string{
			"Coolus Maximus",
			"Cool Dude",
		},
		IsCool: true,
	}

	log.Println(fmt.Sprintf("Original object: %v", cr))

	parameters, err := mapqueryparam.Encode(&cr)
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Parameters: %v", parameters))

	var cr2 *CoolRoot
	err = mapqueryparam.Decode(parameters, &cr2)
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Decoded object: %v", cr2))
}
