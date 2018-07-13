package main

// fridgelethicsdaemon.go provides the Fridgelethics backend daemon that is continuously watching for claim events
// emitted by the Fridgelethics smart contract, verifies the claimable tokens of a user (request to polling service)
// and mints new tokens if applicable.

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	ethProvider = "wss://ropsten.infura.io/ws"                                       // Provider URL for Ethereum RPC API.
	contract    = "b73E516Df09582C043953A66E08E226b34F50711"                         // Contract's address without 0x.
	privKey     = "036d96850f15365ec466915072df4447406da25f6f2c35773976d34561ee3af0" // Fridgelethics owner private key. //TODO do not use for production
)

// main starts the daemon by connecting to an ethereum node, instantiating a contract instance and watching for
// claim events emitted by the contract. For each claim event a PollRequest is sent to the polling service (see
// queryClaimableTokens).
func main() {
	// Set up eth ethClient to make RPC calls to.
	ethClient, err := ethclient.Dial(ethProvider)
	if err != nil {
		log.Fatal("Could not connect to ethereum node:", err)
	}

	// Instantiate contract instance.
	fc, err := NewFridgelethicsClient(ethClient, contract, privKey)
	if err != nil {
		log.Println("Could not instantiate Fridgelethics contract instance:", err)
	}

	// Subscribe to claim events.
	sub, logs := fc.watchClaimEvents()
	if err != nil {
		log.Fatal("Could not subscribe to claim event:", err)
	}

	// Wait for claim events.
	fmt.Println("Watching for claim events...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal("Subscription error:", err)
		case claim := <-logs:
			fmt.Println("Claim event occured:")
			fmt.Println("	Address:", claim.To.Hex())
			fmt.Println("	Value:", claim.Value)

			// Handle the claim event in a new goroutine.
			go fc.handleClaimEvent(claim.To, claim.Value)
		}
	}
}
