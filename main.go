package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk/access/grpc"
)

var ctx = context.Background()

func main() {
	err := godotenv.Load()
	panicIfError(err)

	flowClient, err := grpc.NewClient(grpc.MainnetHost)
	panicIfError(err)

	header, err := flowClient.GetLatestBlockHeader(ctx, true)
	panicIfError(err)

	startingHeight := header.Height - 10
	fmt.Printf("Starting block height: %d\n", startingHeight)

	flowChan, errChan, initErr := flowClient.SubscribeExecutionDataByBlockHeight(ctx, startingHeight)
	panicIfError(initErr)

	go runTxInterpreter()
	go runTxSummarizer()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-flowChan:
			if !ok {
				panic("data is not ok")
			}
			for _, chunk := range data.ExecutionData.ChunkExecutionData {
				for _, transaction := range chunk.Transactions {
					storeNewTx(transaction, data.Height)
				}
			}
		case err, ok := <-errChan:
			if !ok {
				panic("error channel is closed")
			}
			fmt.Printf("~~~ ERROR: %s ~~~\n", err.Error())
		}
	}
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
