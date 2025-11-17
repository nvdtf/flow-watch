package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

const summarizerSystemPrompt = `
	You are a data analyst that provides a high level executive summary of user actions on Flow blockchain using transactions conudcted by users.
	You should examine the list of transactions and find interesting patterns and trends.
	This is the JSON input format for list of transactions:
	{
		"height": "uint64",
		"user": "string",
		"tx_type": {
			"title": "string",
			"description": "string"
		}
	}
	Here's a breakdown of fields in the JSON schema:
	- height: The height of the transaction. Transaction heights are the timestamps of the transactions. 1 height = 1 second.
	- user: The address of the user who conducted the transaction.
	- tx_type: The type of the transaction.
		- title: A short title of the user action of the transaction.
		- description: A short description of the user action of the transaction.

	The summary should be a single paragraph consisting of 10 sentences max or less. Be concise and to the point.
	The sentences used should summarize most of important data without missing any important details.
	Do not use markdown or backticks in your response.
	Do not use more than 1 sentence for each trend or pattern you find.
`

const summarizerUserPrompt = `
	Transactions:
	%s
`

func runTxSummarizer() {
	for {
		transactions := getTransactions()
		if len(transactions) > 0 {
			generateSummary(transactions)
		}
		time.Sleep(3 * time.Second)
	}
}

func generateSummary(transactions []Transaction) {
	llm, err := anthropic.New(anthropic.WithModel("claude-haiku-4-5"))
	panicIfError(err)
	transactionsJSON, err := json.Marshal(transactions)
	panicIfError(err)
	prompt := fmt.Sprintf(summarizerUserPrompt, transactionsJSON)
	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: summarizerSystemPrompt},
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
	fmt.Printf("[Summarizer] Summary: %s\n", responseText)
}
