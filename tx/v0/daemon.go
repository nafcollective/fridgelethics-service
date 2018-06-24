package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/proto"

	"github.com/nafcollective/fridgelethics-service/utils"
)

const (
	ethProvider     = "wss://ropsten.infura.io/ws"                                         // Provider URL for Ethereum RPC API.
	abiPath         = "../../utils/Fridgelethics.abi"                                      // Path to contract ABI.
	contract        = "1CC7791Ce31426F79DdD66995c50aECC101233E5"                           // Contract's address without 0x.
	claimEventTopic = "0x47cee97cb7acd717b3c0aa1435d004cd5b3c8c57d70dbceb4e4458bbd60e39d4" // Topic to filter for claim events only.
)

type ClaimEvent struct {
	To    common.Address
	Value *big.Int
}

func main() {
	// Set up eth client to make RPC calls to.
	client, err := ethclient.Dial(ethProvider)
	if err != nil {
		log.Fatal("Could not connect to ethereum node:", err)
	}

	// Set up query for claim events.
	address := common.HexToAddress(contract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{address},                            // Filter for events of Fridgelethics contract.
		Topics:    [][]common.Hash{{common.HexToHash(claimEventTopic)}}, // Filter for claim events only.
	}

	// Subscribe to claim events.
	ew, err := NewEventWatcher(client, query)
	if err != nil {
		log.Fatal("Could not subscribe:", err)
	}

	// Load contract ABI.
	file, err := ioutil.ReadFile(abiPath)
	contractAbi, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		log.Fatal("Could not get contract ABI:", err)
	}

	// Receive event logs.
	for {
		select {
		case err := <-ew.subscription.Err():
			log.Fatal("Subscription error:", err)
		case eventLog := <-ew.logs:
			var event ClaimEvent
			err = contractAbi.Unpack(&event, "Claim", eventLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Claim event received:")
			fmt.Println("Address:", event.To.Hex())
			fmt.Println("Value:", event.Value)
			go SendPollRequest(event.To.Bytes())
		}
	}
}

func SendPollRequest(address []byte) {
	pr := &pb.PollRequest{
		Address: address,
	}
	msg, err := proto.Marshal(pr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sending message", msg)
	//TODO send msg to polling service
}
