package prompt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// NormalizeJSON normalizes a JSON string by unmarshalling and re-marshalling it.
func NormalizeJSON(jsonStr string) (string, error) {
	var i interface{}
	err := json.Unmarshal([]byte(jsonStr), &i)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CompareJSON compares two JSON strings after normalizing them.
func CompareJSON(jsonStr1, jsonStr2 string) (bool, error) {
	normalizedJSONStr1, err := NormalizeJSON(jsonStr1)
	if err != nil {
		return false, err
	}
	normalizedJSONStr2, err := NormalizeJSON(jsonStr2)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(normalizedJSONStr1, normalizedJSONStr2), nil
}

func TestTo_snake_case(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Test1: HelloWorld to hello_world", "HelloWorld", "hello_world"},
		{"Test2: hello_world remains the same", "hello_world", "hello_world"},
		{"Test3: Hello_world to hello_world", "Hello_world", "hello_world"},
		{"Test4: Hello_World to hello_world", "Hello_World", "hello_world"},
		{"Test5: Hello to hello", "Hello", "hello"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			snake_case_string := To_snake_case(tc.input)
			if snake_case_string != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, snake_case_string)
			}
		})
	}
}

type CoolFact struct {
	Fact               string   `json:"fact"`
	Keywords           []string `json:"keywords"`
	Followup_questions []string `json:"followup_questions"`
}

func TestStruct_to_prompt_schema(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "Basic struct with arrays and string fields",
			input: struct {
				Fact               string
				Keywords           []string
				Followup_questions []string
			}{},
			expected: `
			{
				"fact": "something1",
				"keywords": ["something2", "something3"],
				"followup_questions": ["something4", "something5"]
			}
			`,
		},
		{
			name: "Struct with numbers",
			input: struct {
				Count int
				Ages  []int
			}{},
			expected: `
			{
				"count": 123,
				"ages": [123, 456]
			}
			`,
		},
		{
			name: "Struct with nested structs",
			input: struct {
				StringField string
				NumberField int
				Parent      struct {
					Child struct {
						Grandchildren []struct {
							Name string
							Age  int
						}
					}
				}
			}{},
			expected: `
			{
				"string_field": "something1",
				"number_field": 123,
				"parent": {
					"child": {
						"grandchildren": [
							{
								"name": "something2",
								"age": 123
							},
							{
								"name": "something3",
								"age": 123
							}
						]
					}
				}
			}
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			counter := 1
			string_prompt := Prompt[any, any]{}.Struct_to_prompt_schema(tc.input, &counter)
			isEqual, err := CompareJSON(string_prompt, tc.expected)
			if err != nil {
				t.Errorf("Error while comparing JSON: %s", err)
			}
			if !isEqual {
				t.Errorf("Expected %s, got %s", tc.expected, string_prompt)
			}
		})
	}
}

func TestGenerate_prompt(t *testing.T) {
	type Arguments struct {
		Fact string
	}
	type ChildStruct struct {
		name string
		age  int
	}
	type ParentStruct struct {
		children []ChildStruct
	}

	p := Prompt[ParentStruct, Arguments]{
		Prompt:           "I want {{Fact}} a cool question about {{Fact}}",
		Array_of_results: true,
		Arguments:        Arguments{},
		Json_output:      ParentStruct{},
	}

	prompt := p.Generate_prompt(RunOptions[ParentStruct, Arguments]{
		arguments: Arguments{
			Fact: "something",
		},
	})

	fmt.Println(prompt)
}
