package prompt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

func run_prompt(prompt string, client *openai.Client, streaming_response chan<- string) string {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v", err)
		return ""
	}
	defer stream.Close()

	complete_response := ""

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			os.Exit(1)
		}

		complete_response += response.Choices[0].Delta.Content
		streaming_response <- response.Choices[0].Delta.Content
	}

	close(streaming_response)

	return complete_response
}
