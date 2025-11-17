package main

import (
	"crypto/sha256"
	"fmt"
	"slices"

	"github.com/onflow/flow-go-sdk"
)

type UniqueTransaction struct {
	Count          int                  `json:"-"`
	Description    *InterpreterResponse `json:"description"`
	LastObservedId string               `json:"-"`
	Script         string               `json:"-"`
}

type Transaction struct {
	Height uint64             `json:"height"`
	User   string             `json:"user"`
	TxType *UniqueTransaction `json:"tx_type"`
}

var TX_ID_EXCLUDE_LIST = []string{
	"3408f8b1aa1b33cfc3f78c3f15217272807b14cec4ef64168bcf313bc4174621",
	"a9caece21b073a85cdfa8e27c6781426025ab67d7018b9afe388a18cc293e14f",
}

var TX_HASH_EXCLUDE_LIST = []string{
	"b24cb88aa47956eeef7e2cfde323788968c2a048867e5e2a3d9b4603c5425091", // EVM runner
}

var uniqueTransactions = make(map[[32]byte]*UniqueTransaction)
var transactions = []Transaction{}

func storeNewTx(tx *flow.Transaction, height uint64) {
	if slices.Contains(TX_ID_EXCLUDE_LIST, tx.ID().String()) {
		return
	}
	scriptHash := sha256.Sum256(tx.Script)
	if slices.Contains(TX_HASH_EXCLUDE_LIST, fmt.Sprintf("%x", scriptHash)) {
		return
	}
	utx, exists := uniqueTransactions[scriptHash]
	if exists {
		utx.Count++
		utx.LastObservedId = tx.ID().String()
	} else {
		utx = &UniqueTransaction{
			Count:          1,
			Description:    nil,
			LastObservedId: tx.ID().String(),
			Script:         string(tx.Script),
		}
		uniqueTransactions[scriptHash] = utx
	}
	transactions = append(transactions, Transaction{
		Height: height,
		User:   tx.Authorizers[0].String(),
		TxType: utx,
	})
}

func getTransactions() []Transaction {
	returnList := []Transaction{}
	for _, tx := range transactions {
		if tx.TxType.Description == nil {
			continue
		}
		returnList = append(returnList, tx)
	}
	slices.SortFunc(returnList, func(a, b Transaction) int {
		return int(b.Height - a.Height)
	})
	return returnList
}
