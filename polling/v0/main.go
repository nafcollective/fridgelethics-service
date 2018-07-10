package main // import "github.com/nafcollective/fridgelethics-service/polling/v0"

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	messages "github.com/nafcollective/fridgelethics-messages"
	"google.golang.org/grpc"
)

func main() {
	s := service{}
	ctx := context.Background()
	errc := make(chan error)

	go s.StartGRPCServer(ctx, errc)

	// Interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	<-errc
}

func (s service) StartGRPCServer(ctx context.Context, errc chan error) {
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		errc <- err
		return
	}
	server := grpc.NewServer()

	ps := s.NewGRPCServer(ctx)
	messages.RegisterPollingServiceServer(server, ps)

	errc <- server.Serve(ln)
}
