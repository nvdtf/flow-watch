package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

type InterpreterResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func runTxInterpreter() {
	for {
		for hash, utx := range uniqueTransactions {
			if utx.Count > 1 && utx.Description == nil {
				interpreterResponse := generateDescription(utx.Script)
				// interpreterResponse := mockResponse()
				utx.Description = &interpreterResponse
				uniqueTransactions[hash] = utx
				fmt.Printf("[Interpreter] ID: %s, Count: %d, Description: %+v\n", utx.LastObservedId, utx.Count, interpreterResponse)
			}
		}
		time.Sleep(1 * time.Second)
	}

}

const interpreterSystemPrompt = `
	You are a helpful assistant that explains the user action of Flow blockchain transactions written in Cadence.
	Always respond with valid JSON matching the following schema and nothing else. Do not return any markdown or backticks:
	{
		"title": "string"
		"description": "string"
	}
	Here's a breakdown of fields in the JSON schema:
	- title: A short title of the user action of the transaction.
	- description: A short description of the user action of the transaction.
`

const interpreterUserPrompt = `
	Explain the user action of this Flow blockchain transaction written in Cadence.
	Script:
	%s
`

func mockResponse() InterpreterResponse {
	return InterpreterResponse{
		Title:       "Mock Title",
		Description: "Mock Description",
	}
}

func generateDescription(script string) InterpreterResponse {
	llm, err := anthropic.New(anthropic.WithModel("claude-haiku-4-5"))
	panicIfError(err)
	prompt := fmt.Sprintf(interpreterUserPrompt, script)
	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: interpreterSystemPrompt},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: prompt},
			},
		},
	}
	response, err := llm.GenerateContent(ctx, content)
	panicIfError(err)
	if len(response.Choices) == 0 {
		panic("No response choices returned")
	}
	responseText := response.Choices[0].Content
	var interpreterResponse InterpreterResponse
	err = json.Unmarshal([]byte(cleanJSONResponse(responseText)), &interpreterResponse)
	panicIfError(err)
	return interpreterResponse
}

func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimSuffix(response, "```")
	return response
}
