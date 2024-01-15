package main

import (
	"bytes"
	"log"
	"regexp"
	"strings"
)

func To_snake_case(str string) string {
	// Convert camelcase to snake case
	runes := []rune(str)
	length := len(runes)
	var buffer bytes.Buffer

	for i := 0; i < length; i++ {
		if runes[i] >= 'A' && runes[i] <= 'Z' && i != 0 {
			buffer.WriteString("_")
		}
		buffer.WriteString(string(runes[i]))
	}

	// Convert spaces to underscores
	str = strings.Replace(buffer.String(), " ", "_", -1)
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processed_string := reg.ReplaceAllString(str, "_")
	return strings.ToLower(processed_string)
}
