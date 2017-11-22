package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"context"

	"github.com/DaveAppleton/etherUtils"
	"github.com/DaveAppleton/parityclient"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type mewData struct {
	Address common.Address
	Symbol  string
	Decimal uint8
}

var tokenData []mewData

func main() {
	tokenz := make(map[common.Address]mewData)
	resp, err := http.Get("https://raw.githubusercontent.com/kvhnuke/etherwallet/mercury/app/scripts/tokens/ethTokens.json")
	if err != nil {
		log.Fatal("Cannot load MEW Tokens")
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		log.Fatal("Cannot decode MEW Tokens")
	}
	fmt.Printf("loaded %d tokens from MEW\n", len(tokenData))

	client, err := parityclient.GetClient("http://localhost:8545") //"/Users/daveappleton/Library/Ethereum/geth.ipc") //
	if err != nil {
		log.Fatal("GetClient: ", err)
	}
	fmt.Println("Connected")
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	filter := ethereum.FilterQuery{}
	filter.Addresses = make([]common.Address, 0)
	for _, token := range tokenData {
		filter.Addresses = append(filter.Addresses, token.Address)
		tokenz[token.Address] = token
	}
	filter.Topics = [][]common.Hash{[]common.Hash{transferTopic}}
	filter.Topics = [][]common.Hash{}
	filter.FromBlock = big.NewInt(1000000)
	ctx := context.Background()

	fID, err := client.NewFilter(ctx, filter)
	if err != nil {
		log.Fatal("GetClient: ", err)
	}
	fmt.Println("Filter ID ", fID)

	for {
		logEntries, err := client.FilterChanges(ctx, fID)
		if err != nil {
			log.Fatal("FilterChanges: ", err)
		}
		if len(logEntries) > 0 {

			for _, logEnt := range logEntries {
				if logEnt.Topics[0] != transferTopic {
					continue
				}
				coin := "unknown"
				coinEnt, ok := tokenz[logEnt.Address]
				if ok {
					coin = coinEnt.Symbol
				}
				amount := new(big.Int).SetBytes(logEnt.Data)
				src := common.HexToAddress(logEnt.Topics[1].Hex())
				dest := common.HexToAddress(logEnt.Topics[2].Hex())
				fmt.Printf("%s Transfer(%s,%s,%s)\n", coin, src.Hex(), dest.Hex(), etherUtils.CoinToStr(amount, int(coinEnt.Decimal)))
			}
		} else {
			time.Sleep(time.Second * 5)
		}
	}

}
