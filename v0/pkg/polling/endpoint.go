package polling

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nafcollective/fridgelethics-messages"
)

type Endpoints struct {
	PollEndpoint     endpoint.Endpoint
	RegisterEndpoint endpoint.Endpoint
}

func makePollEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(messages.PollRequest)
		dst, ct := s.Poll(ctx, req.Address)
		return messages.PollResponse{
			Distance:        dst,
			ClaimableTokens: ct,
		}, nil
	}
}

func makeRegisterEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(messages.RegisterRequest)
		url := s.Register(ctx, req.Address)
		return messages.RegisterResponse{
			Url: url,
		}, nil
	}
}

func makeEndpoints(s Service) Endpoints {
	return Endpoints{
		makePollEndpoint(s),
		makeRegisterEndpoint(s),
	}
}
