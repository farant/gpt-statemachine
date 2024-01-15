package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/farant/gpt-statemachine/prompt"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// DONE: arguments to the predefined prompt, how do you require the arguments?
// DONE: JSON structure of output, dynamic return type of argument
// DONE: Include the output struct in the prompt
// DONE: Make unicode characters work ok
// TODO: Make it work with arrays of ints?

func runTimelineEvents(client *openai.Client, subject string) {
	type Event struct {
		YearPublished                string   `json:"year_published"`
		FullNameOfAuthor             string   `json:"full_name_of_author"`
		Title                        string   `json:"title"`
		Description                  string   `json:"description"`
		CounterIntuitivePropositions []string `json:"counter_intuitive_propositions"`
	}

	type Arguments struct {
		Subject string
	}

	timeline_events := prompt.Prompt[Event, Arguments]{
		Prompt: `
		Please generate a timeline of 10 important scientific papers related to {{Subject}}. 
		`,
		Arguments:        Arguments{},
		Json_output:      Event{},
		Array_of_results: true,
	}

	print_events := func(events []Event) {
		// Combine counterintuitive propositions of duplicate papers
		combinedEvents := make(map[string]Event)
		for _, event := range events {
			key := event.Title + event.FullNameOfAuthor
			if val, ok := combinedEvents[key]; ok {
				val.CounterIntuitivePropositions = append(val.CounterIntuitivePropositions, event.CounterIntuitivePropositions...)
				combinedEvents[key] = val
			} else {
				combinedEvents[key] = event
			}
		}
		// Convert the map back to slice
		sortedEvents := make([]Event, 0, len(combinedEvents))
		for _, event := range combinedEvents {
			sortedEvents = append(sortedEvents, event)
		}

		// Sort the slice by year
		sort.Slice(sortedEvents, func(i, j int) bool {
			yearI, _ := strconv.Atoi(sortedEvents[i].YearPublished)
			yearJ, _ := strconv.Atoi(sortedEvents[j].YearPublished)
			return yearI < yearJ
		})

		// Print the combined events
		for _, event := range sortedEvents {
			fmt.Printf("\n- %5s: \"%s\" by %s\n  %s\n", event.YearPublished, event.Title, event.FullNameOfAuthor, event.Description)
			for _, proposition := range event.CounterIntuitivePropositions {
				fmt.Printf("  * %s\n", proposition)
			}
		}
	}

	result := timeline_events.Run(client, prompt.RunOptions[Event, Arguments]{
		Arguments: Arguments{
			Subject: subject,
		},
		On_json_array_progress: func(progress []Event, raw_string string) {
			fmt.Print("\033[H\033[2J")
			print_events(progress)
		},
	})

	fmt.Printf("PROMPT:\n\n")
	fmt.Println(result.Prompt_text)
	fmt.Printf("\n\nRESPONSE:\n\n")
	fmt.Println(result.Parsed_results_json)
	fmt.Printf("\n\nPARSED:\n\n")
	print_events(result.Parsed_results_array)
}

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

	runTimelineEvents(client, combined_args)
}
