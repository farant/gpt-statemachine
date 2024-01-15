package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/farant/gpt-statemachine/besteffortjson"
	"github.com/sashabaranov/go-openai"
)

type Prompt[Output any, Input any] struct {
	Prompt           string
	Json_output      Output
	Array_of_results bool
	Arguments        Input
}

type RunOptions[Output any, Input any] struct {
	on_response_progress   func(string)
	on_json_array_progress func([]Output)
	arguments              Input
}

func (p Prompt[Output, Input]) Run(client *openai.Client, options RunOptions[Output, Input]) string {
	streaming_response := make(chan string)

	/*
		if options.on_response_progress != nil {
			go func() {
				for response := range streaming_response {
					options.on_response_progress(response)
				}
			}()
		}
	*/

	if options.on_json_array_progress != nil {
		if !p.Array_of_results {
			panic(fmt.Errorf("on_json_array_progress requires Array_of_results to be true"))
		}

		go func() {
			total_progress := ""
			for response := range streaming_response {
				total_progress += response
				result_json := besteffortjson.Best_effort_json_parse(total_progress)

				var response struct {
					Results []Output `json:"results"`
				}

				err := json.Unmarshal([]byte(result_json), &response)
				if err != nil {
					log.Println("JSON parse error: ", err)
					return
				}

				options.on_json_array_progress(response.Results)
			}
		}()
	}

	prompt := p.Generate_prompt(options)

	return run_prompt(prompt, client, streaming_response)
}

func (p Prompt[Output, Input]) StructToMap(obj interface{}) map[string]interface{} {
	v := reflect.ValueOf(obj)
	out := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Name
		value := v.Field(i).Interface()

		out[key] = value
	}

	return out
}

func (p Prompt[Output, Input]) Generate_prompt(options RunOptions[Output, Input]) string {
	prompt := p.Prompt
	re := regexp.MustCompile(`{{(\w+)}}`)
	matches := re.FindAllStringSubmatch(prompt, -1)
	arguments_map := p.StructToMap(options.arguments)
	for _, match := range matches {
		keyword := match[1]
		if val, ok := arguments_map[match[1]]; ok {
			prompt = strings.Replace(prompt, match[0], fmt.Sprintf("%v", val), -1)
		} else {
			panic(fmt.Errorf("argument not found in options %s", keyword))
		}
	}

	prompt += "\n\n"
	if p.Array_of_results {
		prompt += "In your response send me an array of JSON objects. Don't include any markdown block syntax."
		prompt += "Here's an example result to match:\n\n"
		counter := 1
		prompt += strings.TrimSpace(fmt.Sprintf(`
{
	"results": [
		%s,
		// etc.
	]
		}
			`, p.Struct_to_prompt_schema(p.Json_output, &counter)))
	} else {
		prompt += "In your response send me a JSON object. Don't include any markdown block syntax.\n"
		prompt += "Here's an example result to match:\n\n"
		counter := 1
		prompt += p.Struct_to_prompt_schema(p.Json_output, &counter)
	}

	return prompt
}

func (p Prompt[Output, Input]) Struct_to_prompt_schema(struct_type interface{}, something_counter *int) string {
	v := reflect.ValueOf(struct_type)
	typeOfS := v.Type()

	next_something := func() string {
		result := "something" + fmt.Sprintf("%d", *something_counter)
		(*something_counter)++
		return result
	}

	output := "{\n"
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		var example_value string
		switch field.Kind() {
		case reflect.Slice:
			elemType := field.Type().Elem()
			if elemType.Kind() == reflect.Int {
				example_value = "[123, 456]"
			} else if elemType.Kind() == reflect.Struct {
				example_value = fmt.Sprintf("[%s, %s]", p.Struct_to_prompt_schema(reflect.New(elemType).Elem().Interface(), something_counter), p.Struct_to_prompt_schema(reflect.New(elemType).Elem().Interface(), something_counter))
			} else {
				example_value = fmt.Sprintf("[\"%s\", \"%s\"]", next_something(), next_something())
			}
		case reflect.Struct:
			example_value = p.Struct_to_prompt_schema(field.Interface(), something_counter)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			example_value = "123"
		case reflect.Float32, reflect.Float64:
			example_value = "123.0"
		default:
			example_value = fmt.Sprintf("\"%s\"", next_something())
		}

		if i != v.NumField()-1 {
			output += fmt.Sprintf(
				"\t\"%s\": %s,\n",
				To_snake_case(typeOfS.Field(i).Name),
				example_value,
			)
		} else {
			output += fmt.Sprintf(
				"\t\"%s\": %s",
				To_snake_case(typeOfS.Field(i).Name),
				example_value,
			)
		}
	}

	output += "\n}"

	return output
}
