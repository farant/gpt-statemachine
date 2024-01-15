package besteffortjson

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

func parse_value(current_index int, json_string string) (interface{}, int) {
	var result interface{}

	final_index := current_index
	var next_index int

	for i := current_index; i < len(json_string); i++ {
		char := rune(json_string[i])
		final_index = i
		if char == ' ' ||
			char == '\n' ||
			char == '\t' ||
			char == '}' ||
			char == ',' ||
			char == ']' {
			continue
		}

		switch char {
		case '{':
			result, next_index = parse_object(i, json_string)
		case '[':
			result, next_index = parse_array(i, json_string)
		case '"':
			result, next_index = parse_string(i, json_string)
		default:
			result, next_index = parse_number_or_boolean_or_null(i, json_string)
		}
		i = next_index
		final_index = i
		break
	}

	return result, final_index
}

func parse_object(current_index int, json_string string) (map[string]interface{}, int) {
	result := make(map[string]interface{})
	final_index := current_index

	state := "looking_for_key"
	current_key := ""
	var previous_char rune
	for i := current_index; i < len(json_string); i++ {
		char := rune(json_string[i])
		final_index = i

		if char == '}' {
			break
		}

		switch state {
		case "looking_for_key":
			if char == '"' {
				state = "in_key"
				current_key = ""
			}
		case "in_key":
			if char == '"' && previous_char != '\\' {
				state = "looking_for_colon"
			} else {
				current_key += string(char)
			}
		case "looking_for_colon":
			if char == ':' {
				state = "looking_for_value"
			}
		case "looking_for_value":
			result[current_key], i = parse_value(i, json_string)
			state = "looking_for_key"
		}

		previous_char = char
		final_index = i
	}

	return result, final_index
}

func parse_array(current_index int, json_string string) ([]interface{}, int) {
	var result []interface{}
	final_index := current_index

	found_value := false
	for i := final_index; i < len(json_string); i++ {
		char := rune(json_string[i])
		if char == ']' {
			final_index = i
			return result, final_index
		} else if char == ',' || char == ' ' || char == '\n' || char == '\t' {
			final_index++
		} else if char == '[' && !found_value {
			final_index++
		} else {
			found_value = true
			value, next_index := parse_value(i, json_string)
			result = append(result, value)
			i = next_index
		}

		final_index = i
	}

	return result, final_index
}

func parse_string(current_index int, json_string string) (string, int) {
	var result string
	final_index := current_index
	state := "starting_string"
	for i := current_index; i < len(json_string); i++ {
		char := rune(json_string[i])
		switch state {
		case "starting_string":
			if char == '"' {
				state = "in_string"
			}
		case "found_slash":
			if char == '"' {
				result += `"`
			} else if char == 'n' {
				result += "\n"
			} else if char == 't' {
				result += "\t"
			} else if char == '\\' {
				result += "\\"
			} else {
				result += "\\" + string(char)
			}

			state = "in_string"
		case "in_string":
			if char == '\\' {
				state = "found_slash"
			} else if char == '"' {
				return result, i
			} else {
				result += string(char)
			}
		}

		final_index = i
	}

	return result, final_index
}

func parse_number_or_boolean_or_null(current_index int, json_string string) (interface{}, int) {
	var result interface{}
	final_index := current_index

	var literal_type string
	var numeric_string string

	var found_value = false
	for _, char := range json_string[current_index:] {
		final_index++

		if found_value {
			if char == ',' ||
				char == '}' ||
				char == ']' ||
				char == ' ' ||
				char == '\n' ||
				char == '\t' {
				final_index--
				break
			}
		} else {
			if char == ' ' || char == '\n' || char == '\t' {
				continue
			}
		}

		found_value = true
		switch {
		case (char >= '0' && char <= '9') || char == '-' || char == '.':
			literal_type = "number"
			numeric_string += string(char)
		case char == 't' || char == 'f' || literal_type == "boolean":
			literal_type = "boolean"
			if result == nil {
				if char == 't' {
					result = true
				} else {
					result = false
				}
			}
		case char == 'n' || literal_type == "null":
			literal_type = "null"
		}
	}

	if literal_type == "number" {
		if strings.Contains(numeric_string, ".") {
			result, _ = strconv.ParseFloat(numeric_string, 64)
		} else {
			tempResult, _ := strconv.ParseInt(numeric_string, 10, 0)
			result = int(tempResult)
		}
	}

	return result, final_index
}

func Best_effort_json_parse(in_progress string) string {
	lines := strings.Split(in_progress, "\n")
	new_in_progress := ""
	found_beginning := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			found_beginning = true
		}
		if found_beginning && strings.HasPrefix(strings.TrimSpace(line), "```") {
			break
		}
		if found_beginning {
			new_in_progress += line
		}
	}
	in_progress = new_in_progress

	result, _ := parse_value(0, in_progress)

	in_progress_result, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	return string(in_progress_result)
}
