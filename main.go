package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// DONE: arguments to the predefined prompt, how do you require the arguments?
// DONE: JSON structure of output, dynamic return type of argument
// DONE: Include the output struct in the prompt
// DONE: Make unicode characters work ok
// TODO: Make it work with arrays of ints?

func main() {
	args := os.Args[1:]
	combined_args := strings.Join(args, " ")

	if combined_args == "" {
		fmt.Println("Error: No arguments provided.")
		os.Exit(1)
	}

	godotenv.Load()

	api_key := os.Getenv("OPENAI_API_KEY")
	if api_key == "" {
		panic("please set the OPENAI_API_KEY environment variable")
	}

	client := openai.NewClient(api_key)

	type Place struct {
		Name        string `json:"name"`
		Coordinates struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"coordinates"`
		FamousPeople []struct {
			Name        string `json:"name"`
			YearOfBirth string `json:"year_of_birth"`
			YearOfDeath string `json:"year_of_death"`
			Events      []struct {
				Description string `json:"description"`
				Year        string `json:"year"`
			}
		} `json:"famous_people"`
	}

	type Arguments struct {
		Region string
	}

	historical_events := Prompt[Place, Arguments]{
		Prompt: `
		Please tell me about 6 locations in {{Region}}.

		Along with each location tell me about the people who are famous and connected to the place
		with a few historical events related to that person (other than birth and death).
		`,
		Arguments:        Arguments{},
		Json_output:      Place{},
		Array_of_results: true,
	}

	result := historical_events.Run(client, RunOptions[Place, Arguments]{
		arguments: Arguments{
			Region: combined_args,
		},
		on_json_array_progress: func(progress []Place, raw_string string) {
			fmt.Print("\033[H\033[2J")

			for _, place := range progress {
				fmt.Printf("\n%s (%f, %f)\n", place.Name, place.Coordinates.Latitude, place.Coordinates.Longitude)
				fmt.Println("\nFamous people:")
				for _, person := range place.FamousPeople {
					fmt.Printf("\n- %s (%s - %s)\n", person.Name, person.YearOfBirth, person.YearOfDeath)
					fmt.Println("  Events:")
					for _, event := range person.Events {
						fmt.Printf("  - %s (%s)\n", event.Description, event.Year)
					}
				}
			}
		},
	})

	fmt.Printf("PROMPT:\n\n")
	fmt.Println(result.prompt_text)
	fmt.Printf("\n\nRESPONSE:\n\n")
	fmt.Println(result.parsed_results_json)
}
