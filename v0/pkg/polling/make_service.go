package polling

import (
	"context"
	"net"

	"github.com/go-redis/redis"
	messages "github.com/nafcollective/fridgelethics-messages"
	"google.golang.org/grpc"
)

func NewService(ctx context.Context, o *redis.Options, errc chan error) (Service, error) {
	s := &service{
		db: redis.NewClient(o),
		fs: nil,
	}

	_, err := s.db.Ping().Result()
	if err != nil {
		return nil, err
	}

	endpoints := makeEndpoints(s)
	ps := newGRPCServer(ctx, endpoints)
	go s.startGRPCServer(ps, errc)

	return s, nil
}

func (s service) Stop() error {
	//TODO: Write Stop function
	return nil
}

func (s service) startGRPCServer(ps messages.PollingServiceServer, errc chan error) {
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		errc <- err
		return
	}
	server := grpc.NewServer()

	messages.RegisterPollingServiceServer(server, ps)

	errc <- server.Serve(ln)
}
