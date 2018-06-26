package main

// fridgelethicsclient.go is a wrapper for the relevant functions from fridgelethics.go (generated using the abigen
// tool). It provides the essential functionality for the Fridgelethics backend to interact with the Fridgelethics
// smart contract such as subscribing to claim events and sending Mint() transactions.

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/nafcollective/fridgelethics-service/tx/v0/abigen"
)

// FridgelethicsClient is a wrapper struct to interact with the Fridgelethics smart contract.
type FridgelethicsClient struct {
	contractClient *abigen.Fridgelethics // Instance of the Fridgelethics smart contract.
	auth           *bind.TransactOpts    // Authorization to interact with the contract.
}

// NewFridgelethicsClient creates a new FridgelethicsClient using the ethClient as RPC API. contract is the contract
// address (without '0x') and privKey should be the private key of the contract's owner.
func NewFridgelethicsClient(ethClient *ethclient.Client, contract, privKey string) (fc *FridgelethicsClient, err error) {
	// Instantiate contract instance.
	address := common.HexToAddress(contract)
	f, err := abigen.NewFridgelethics(address, ethClient)

	// Authorize to access contract.
	key, err := crypto.HexToECDSA(privKey)
	auth := bind.NewKeyedTransactor(key)

	return &FridgelethicsClient{
		contractClient: f,
		auth:           auth,
	}, err
}

// WatchClaimEvents subscribes to claim events emitted by the Fridgelethics smart contract. It returns a sub to catch
// errors or unsubscribe as well as logs, a channel to receive the events from.
func (fc FridgelethicsClient) WatchClaimEvents() (sub ethereum.Subscription, logs chan *abigen.FridgelethicsClaim) {
	logs = make(chan *abigen.FridgelethicsClaim)
	sub, err := fc.contractClient.FridgelethicsFilterer.WatchClaim(nil, logs)
	if err != nil {
		log.Fatal("Could not subscribe to claim events:", err)
	}
	return sub, logs
}

// Mint sends a Mint() transaction that mints the provided amount of Fridgelethics tokens and credits them to the
// address provided.
func (fc FridgelethicsClient) Mint(to common.Address, amount *big.Int) (tx *types.Transaction, err error) {
	return fc.contractClient.Mint(fc.auth, to, amount)
}
