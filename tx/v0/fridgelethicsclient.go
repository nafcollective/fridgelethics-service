package main

// fridgelethicsclient.go is a wrapper for the relevant functions from fridgelethics.go (generated using the abigen
// tool). It provides the essential functionality for the Fridgelethics backend to interact with the Fridgelethics
// smart contract such as subscribing to claim events and sending Mint() transactions.

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/proto"

	messages "github.com/nafcollective/fridgelethics-messages"
)

// FridgelethicsClient is a wrapper struct to interact with the Fridgelethics smart contract.
type FridgelethicsClient struct {
	ethClient      *ethclient.Client
	contractClient *Fridgelethics     // Instance of the Fridgelethics smart contract.
	auth           *bind.TransactOpts // Authorization to interact with the contract.
}

// NewFridgelethicsClient creates and returns a new FridgelethicsClient using the ethClient as RPC API. contract is
// the contract address (without '0x') and privKey should be the private key of the contract's owner.
func NewFridgelethicsClient(ethClient *ethclient.Client, contract, privKey string) (fc *FridgelethicsClient, err error) {
	// Instantiate contract instance.
	address := common.HexToAddress(contract)
	f, err := NewFridgelethics(address, ethClient)

	// Authorize to access contract.
	key, err := crypto.HexToECDSA(privKey)
	auth := bind.NewKeyedTransactor(key)

	return &FridgelethicsClient{
		ethClient:      ethClient,
		contractClient: f,
		auth:           auth,
	}, err
}

// handleClaimEvent gets called from the daemon when a new claim event was caught. It determines the
// gas price for the mint transaction, queries the polling service for the amount of claimable tokens
// of the sender and finally mints the new Fridgelethics tokens and credits them to the sender.
func (fc FridgelethicsClient) handleClaimEvent(to common.Address, weiSent *big.Int) {
	// Determine gas price to use for mint transaction depending on eth sent by the user.
	gasPrice, err := fc.determineMintGasPrice(to, weiSent)
	if err != nil {
		log.Fatal("Could not determine gas price:", err)
		return
	}
	// Check if gasPrice is higher than one.
	if gasPrice.Cmp(big.NewInt(1)) < 0 {
		log.Fatal("User did not send enough eth to fund mint transaction")
		return
	}

	// Query polling service for claimable tokens of that user.
	amount, err := queryClaimableTokens(to.Bytes())
	if err != nil {
		log.Fatal("Could not query for claimable tokens:", err)
		return
	}

	// Mint tokens.
	_, err = fc.mint(to, big.NewInt(int64(amount)))
	if err != nil {
		log.Fatal("Could not mint tokens:", err)
		return
	}
	return
}

// watchClaimEvents subscribes to claim events emitted by the Fridgelethics smart contract. It returns a sub to catch
// errors or unsubscribe as well as logs, a channel to receive the events from.
func (fc FridgelethicsClient) watchClaimEvents() (sub geth.Subscription, logs chan *FridgelethicsClaim) {
	logs = make(chan *FridgelethicsClaim)
	sub, err := fc.contractClient.FridgelethicsFilterer.WatchClaim(nil, logs)
	if err != nil {
		log.Fatal("Could not subscribe to claim events:", err)
	}
	return sub, logs
}

// determineMintGasPrice determines and returns the gas price in gwei that should be used for the mint transaction
// based on the amount of wei sent by the user when claiming tokens.
func (fc FridgelethicsClient) determineMintGasPrice(to common.Address, weiSent *big.Int) (gasPrice *big.Int, err error) {
	// Estimate gas used.
	gas, err := fc.determineMintGas(context.Background(), to)
	if err != nil {
		return
	}

	// Convert wei to gwei.
	gweiSent := big.NewInt(0)
	gweiSent.Div(weiSent, big.NewInt(1000000000))

	// Determine gas price.
	gasPrice = big.NewInt(0)
	gasPrice.Div(gweiSent, gas)
	return
}

// determineMintGas estimates and returns the amount of gas used for minting Fridgelethics tokens to the
// address specified. The amount of tokens is hardcoded to a default as it does not have an influence gas usage.
func (fc FridgelethicsClient) determineMintGas(ctx context.Context, to common.Address) (*big.Int, error) {
	// Encode mint function parameters using the ABI.
	abi, err := abi.JSON(strings.NewReader(FridgelethicsABI))
	if err != nil {
		return nil, err
	}
	data, err := abi.Pack("mint", to, big.NewInt(int64(10))) // the amount of claimed tokens does not have influence on gas
	if err != nil {
		return nil, err
	}

	// Create the message and estimate the gas for its execution.
	scAddr := common.HexToAddress("0x" + contract)
	msg := geth.CallMsg{
		From: fc.auth.From,
		To:   &scAddr,
		Data: data,
	}
	gas, err := fc.ethClient.EstimateGas(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Return big.Int.
	gasBig := big.NewInt(int64(gas))
	return gasBig, err
}

// mint sends and returns a mint() transaction that mints the provided amount of Fridgelethics tokens and credits them to the
// address provided.
func (fc FridgelethicsClient) mint(to common.Address, amount *big.Int) (tx *types.Transaction, err error) {
	return fc.contractClient.Mint(fc.auth, to, amount)
}

// queryClaimableTokens packs the address received in a claim event into a PollRequest and sends it to the polling
// service to verify tokens can be claimed by that address. It returns the number of claimable tokens.
func queryClaimableTokens(address []byte) (tokens uint, err error) {
	pr := &messages.PollRequest{
		Address: address,
	}
	msg, err := proto.Marshal(pr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sending message to polling service", msg)
	//TODO send msg to polling service
	//TODO receive answer from polling service
	//TODO return claimable tokens
	return 5, nil
}
