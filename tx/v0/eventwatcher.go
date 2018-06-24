package main

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EventWatcher subscribes to events according to the provided filters.
type EventWatcher struct {
	ethClient    ethclient.Client      // Ethereum RPC API.
	subscription ethereum.Subscription // Subscription to events according to provided filters.
	logs         chan types.Log        // Channel to receive events.
}

// NewEventWatcher starts the event subscription according to the filterQuery and returns a new EventWatcher.
func NewEventWatcher(ethClient *ethclient.Client, filterQuery ethereum.FilterQuery) (*EventWatcher, error) {
	// Set up subscription.
	var ch = make(chan types.Log)
	ctx := context.Background()
	sub, err := ethClient.SubscribeFilterLogs(ctx, filterQuery, ch)

	// Return new EventWatcher.
	return &EventWatcher{
		ethClient:    *ethClient,
		subscription: sub,
		logs:         ch,
	}, err
}
