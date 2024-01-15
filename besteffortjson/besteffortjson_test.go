package besteffortjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParse_string(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello`, "hello"},
		{`"hel\"lo`, `hel"lo`},
		{`"hel\nlo`, "hel\nlo"},
		{`"hel\\nlo`, "hel\\nlo"},
	}

	for _, tc := range testCases {
		result, _ := parse_string(0, tc.input)
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}

func TestParse_number_or_boolean_or_null(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{`true`, true},
		{`false`, false},
		{`null`, nil},
		{`123`, 123},
		{`123.456`, 123.456},
		{`-123.456`, -123.456},
		{`0.456`, 0.456},
		{`-0.456`, -0.456},
		{`0`, 0},
		{`-0`, 0},
	}
	for _, tc := range testCases {
		result, _ := parse_number_or_boolean_or_null(0, tc.input)
		if result != tc.expected {
			resultType := reflect.TypeOf(result)
			expectedType := reflect.TypeOf(tc.expected)
			fmt.Printf("Result type: %s, Expected type: %s\n", resultType, expectedType)
			t.Errorf("Expected %v, got %v", tc.expected, result)
		}
	}
}

func TestParse_array(t *testing.T) {
	testCases := []struct {
		input    string
		expected []interface{}
	}{
		{`[ 1 , 2 , 3 ]`, []interface{}{1, 2, 3}},
		{`[ 1 , 2234 , 3 ]`, []interface{}{1, 2234, 3}},
		{`[1,2,3]`, []interface{}{1, 2, 3}},
		{`[1,2,3,]`, []interface{}{1, 2, 3}},
		{`[1,"2"]`, []interface{}{1, "2"}},
		{`[1, 2.1  , "123`, []interface{}{1, 2.1, "123"}},
		{`["abc", "123]`, []interface{}{"abc", "123]"}},
		{`[   true, false ]`, []interface{}{true, false}},
		{`[   true, false, t`, []interface{}{true, false, true}},
		{`[   null, false, n `, []interface{}{nil, false, nil}},
		{`["abc", ["123`, []interface{}{"abc", []interface{}{"123"}}},
		{
			`[ { "name": "jim" }, { "name": "cathy" }, { "name": "george`,
			[]interface{}{
				map[string]interface{}{"name": "jim"},
				map[string]interface{}{"name": "cathy"},
				map[string]interface{}{"name": "george"},
			},
		},
	}
	for _, tc := range testCases {
		result, _ := parse_array(0, tc.input)
		if !reflect.DeepEqual(result, tc.expected) {
			expectedJson, _ := json.Marshal(tc.expected)
			resultJson, _ := json.Marshal(result)
			t.Errorf("Expected %s, got %s", string(expectedJson), string(resultJson))
		}
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
		inProgress := Best_effort_json_parse(tc.inProgressStr)
		if inProgress != tc.expected {
			t.Errorf("\n\033[32m%s\033[0m\nExpected %v, got %v", tc.name, tc.expected, inProgress)
		}
	}
}
