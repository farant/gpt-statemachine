package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestTo_snake_case(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello_world"},
		{"hello_world", "hello_world"},
		{"Hello_world", "hello_world"},
		{"Hello_World", "hello_world"},
		{"Hello", "hello"},
	}

	for _, tc := range testCases {
		snake_case_string := To_snake_case(tc.input)
		if snake_case_string != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, snake_case_string)
		}
	}
}

func TestStruct_to_prompt_schema(t *testing.T) {
	var question CoolFact
	string_prompt := Struct_to_prompt_schema(question)

	expected_prompt := strings.TrimSpace(`
{
	"fact": "something1",
	"keywords": ["something2", "something3"],
	"followup_questions": ["something4", "something5"]
}
`)

	if string_prompt != expected_prompt {
		t.Errorf("Expected %s, got %s", expected_prompt, string_prompt)
	}
}

func TestGenerate_prompt(t *testing.T) {
	options := RunOptions{
		json_output: CoolFact{},
		arguments: map[string]string{
			"fact": "go subroutines",
		},
	}

	prompt := Generate_prompt("I want {{fact}} a cool question about {{fact}}", options)

	fmt.Println(prompt)
}
