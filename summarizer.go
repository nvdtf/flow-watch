package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

const summarizerSystemPrompt = `
	You are a data analyst that summarizes the user action of Flow blockchain transactions written in Cadence.
	You should examine the list of transactions and find interesting patterns and trends.
	Transaction heights are the timestamps of the transactions. 1 height = 1 second.
	The summary should be a single paragraph consisting of 3 sentences max.
	The sentences used should summarize most of important data without missing any important details.
	Do not use markdown or backticks in your response.
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
		time.Sleep(10 * time.Second)
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
