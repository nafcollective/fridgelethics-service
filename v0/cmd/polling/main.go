package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis"
	polling "github.com/nafcollective/fridgelethics-service/v0/pkg/polling"
)

func main() {
	ctx := context.Background()
	errc := make(chan error)

	_, err := polling.NewService(ctx, &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}, errc)
	if err != nil {
		panic("Couldn't create the polling service: " + err.Error())
	}

	// Interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	print((<-errc).Error())
}
