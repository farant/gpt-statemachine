package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// DONE: arguments to the predefined prompt, how do you require the arguments?
// DONE: JSON structure of output, dynamic return type of argument
// TODO: Make unicode characters work ok
// TODO: Make it work with arrays of ints?
// TODO: Include the output struct in the prompt

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

	type HistoricalEvent struct {
		Description string   `json:"description"`
		Year        string   `json:"year"`
		People      []string `json:"people"`
	}

	type Arguments struct {
		Subject string
	}

	historical_events := Prompt[HistoricalEvent, Arguments]{
		Prompt: `
		Let me know about ten surprising historical events related to the history of {{Subject}}.
		Please avoid any non-ASCII characters in the output. Use negative integers for BC & positive for AD but still send the number as a string.
		`,
		Arguments:        Arguments{},
		Json_output:      HistoricalEvent{},
		Array_of_results: true,
	}

	historical_events.Run(client, RunOptions[HistoricalEvent, Arguments]{
		arguments: Arguments{
			Subject: combined_args,
		},
		on_json_array_progress: func(progress []HistoricalEvent) {
			fmt.Print("\033[H\033[2J")

			events := []string{}
			people := []string{}

			sort.Slice(progress, func(i, j int) bool {
				yearI, errI := strconv.Atoi(progress[i].Year)
				yearJ, errJ := strconv.Atoi(progress[j].Year)
				if errI != nil || errJ != nil {
					return progress[i].Year < progress[j].Year
				}
				return yearI < yearJ
			})

			for _, event := range progress {
				events = append(events, fmt.Sprintf("%5s: %s", event.Year, event.Description))

				people = append(people, event.People...)
			}

			//sort.Strings(events)

			fmt.Println(strings.Join(events, "\n"))

			people = Deduplicate(people)
			sort.Strings(people)
			fmt.Println("\nPeople involved:")
			for _, person := range people {
				fmt.Println("- " + person)
			}
		},
	})
}
