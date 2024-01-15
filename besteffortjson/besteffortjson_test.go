package besteffortjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParse_string(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Basic string", input: `"hello"`, expected: "hello"},
		{name: "String without closing quote", input: `"hello`, expected: "hello"},
		{name: "String with escaped quote", input: `"hel\"lo`, expected: `hel"lo`},
		{name: "String with newline", input: `"hel\nlo`, expected: "hel\nlo"},
		{name: "String with escaped newline", input: `"hel\\nlo`, expected: "hel\\nlo"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := parse_string(0, tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestParse_number_or_boolean_or_null(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{name: "Parse true", input: `true`, expected: true},
		{name: "Parse false", input: `false`, expected: false},
		{name: "Parse null", input: `null`, expected: nil},
		{name: "Parse positive integer", input: `123`, expected: 123},
		{name: "Parse positive float", input: `123.456`, expected: 123.456},
		{name: "Parse negative float", input: `-123.456`, expected: -123.456},
		{name: "Parse positive float less than 1", input: `0.456`, expected: 0.456},
		{name: "Parse negative float less than 1", input: `-0.456`, expected: -0.456},
		{name: "Parse zero", input: `0`, expected: 0},
		{name: "Parse negative zero", input: `-0`, expected: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := parse_number_or_boolean_or_null(0, tc.input)
			if result != tc.expected {
				resultType := reflect.TypeOf(result)
				expectedType := reflect.TypeOf(tc.expected)
				fmt.Printf("Result type: %s, Expected type: %s\n", resultType, expectedType)
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParse_array(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{name: "Array with integers", input: `[ 1 , 2 , 3 ]`, expected: []interface{}{1, 2, 3}},
		{name: "Array with large integer", input: `[ 1 , 2234 , 3 ]`, expected: []interface{}{1, 2234, 3}},
		{name: "Array without spaces", input: `[1,2,3]`, expected: []interface{}{1, 2, 3}},
		{name: "Array with trailing comma", input: `[1,2,3,]`, expected: []interface{}{1, 2, 3}},
		{name: "Array with string", input: `[1,"2"]`, expected: []interface{}{1, "2"}},
		{name: "Array with incomplete string", input: `[1, 2.1  , "123`, expected: []interface{}{1, 2.1, "123"}},
		{name: "Array with incomplete string and bracket", input: `["abc", "123]`, expected: []interface{}{"abc", "123]"}},
		{name: "Array with booleans", input: `[   true, false ]`, expected: []interface{}{true, false}},
		{name: "Array with incomplete boolean", input: `[   true, false, t`, expected: []interface{}{true, false, true}},
		{name: "Array with null and incomplete null", input: `[   null, false, n `, expected: []interface{}{nil, false, nil}},
		{name: "Array with nested array", input: `["abc", ["123`, expected: []interface{}{"abc", []interface{}{"123"}}},
		{
			name:  "Array with objects",
			input: `[ { "name": "jim" }, { "name": "cathy" }, { "name": "george`,
			expected: []interface{}{
				map[string]interface{}{"name": "jim"},
				map[string]interface{}{"name": "cathy"},
				map[string]interface{}{"name": "george"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := parse_array(0, tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				expectedJson, _ := json.Marshal(tc.expected)
				resultJson, _ := json.Marshal(result)
				t.Errorf("Expected %s, got %s", string(expectedJson), string(resultJson))
			}
		})
	}
}

func TestBest_effort_json_parse(t *testing.T) {
	testCases := []struct {
		name          string
		inProgressStr string
		expected      string
	}{
		{
			"partial key doesn't get added",
			`{ "fac `,
			`{}`,
		},
		{
			"key with no value doesn't get added",
			`{ "fact" `,
			`{}`,
		},
		{
			"key with no value and colon still doesn't get added",
			`{ "fact": `,
			`{"fact":null}`,
		},
		{
			"key with open quote is empty string",
			`{ "fact": "`,
			`{"fact":""}`,
		},
		{
			"key with incomplete value is string",
			`{ "fact": "some`,
			`{"fact":"some"}`,
		},
		{
			"can capture values and ignore incomplete keys",
			`{ "fact": "something", "key`,
			`{"fact":"something"}`,
		},
		{
			"one captured value and last incomplete value is a space",
			`{ "fact": "something", "keywords": " `,
			`{"fact":"something","keywords":" "}`,
		},
		{
			"two values, one incomplete is still captured",
			`{ "fact": "something", "keywords": "pizza and such`,
			`{"fact":"something","keywords":"pizza and such"}`,
		},
		{
			"nested objects with incomplete value",
			`{ "fact": { "one": "two`,
			`{"fact":{"one":"two"}}`,
		},
		{
			"nested arrays with incomplete value",
			`{ "fact": [ "one", "two`,
			`{"fact":["one","two"]}`,
		},
		{
			"more complicated result",
			`{
					"fact": [ "one", "two"],
					"results": [
						{
							"name": "john",
							"country
				`,
			`{"fact":["one","two"],"results":[{"name":"john"}]}`,
		},
		{
			"with markdown and extra answer stuff",
			`
				Hello, this is my answer. Very good. etc.

				` + "```" + `json
				{
					"fact": [ "one", "two"],
					"results": [
						{
							"name": "john",
							"country
				`,
			`{"fact":["one","two"],"results":[{"name":"john"}]}`,
		},
		{
			"with markdown and complete answer",
			`
			Hello, this is my answer. Very good. etc.

			` + "```" + `json
			{ 
				"fact": [ "one", "two"],
				"results": [
					{
						"name": "john",
						"country": "usa"
					}
				]
			}
			` + "```" + `
			
			Is that the answer you wanted? etc.
			`,
			`{"fact":["one","two"],"results":[{"country":"usa","name":"john"}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inProgress := Best_effort_json_parse(tc.inProgressStr)
			if inProgress != tc.expected {
				t.Errorf("\n\033[32m%s\033[0m\nExpected %v, got %v", tc.name, tc.expected, inProgress)
			}
		})
	}
}
