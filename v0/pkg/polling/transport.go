package polling

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/nafcollective/fridgelethics-messages"
	oldcontext "golang.org/x/net/context"
)

type grpcServer struct {
	poll     grpctransport.Handler
	register grpctransport.Handler
}

func newGRPCServer(ctx context.Context, endpoints Endpoints) messages.PollingServiceServer {
	return &grpcServer{
		poll: grpctransport.NewServer(
			endpoints.PollEndpoint,
			DecodeGRPCPollRequest,
			EncodeGRPCPollResponse,
		),
		register: grpctransport.NewServer(
			endpoints.RegisterEndpoint,
			DecodeGRPCRegisterRequest,
			EncodeGRPCRegisterResponse,
		),
	}
}

func (s *grpcServer) Poll(ctx oldcontext.Context, req *messages.PollRequest) (*messages.PollResponse, error) {
	_, res, err := s.poll.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.(*messages.PollResponse), nil
}

func (s *grpcServer) Register(ctx oldcontext.Context, req *messages.RegisterRequest) (*messages.RegisterResponse, error) {
	_, res, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.(*messages.RegisterResponse), nil
}

func DecodeGRPCPollRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*messages.PollRequest)
	return messages.PollRequest{
		Address: req.Address,
	}, nil
}

func DecodeGRPCRegisterRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*messages.RegisterRequest)
	return messages.RegisterRequest{
		Address: req.Address,
	}, nil
}

func EncodeGRPCPollResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(messages.PollResponse)
	grpcRes := &messages.PollResponse{
		Distance:        uint64(res.Distance),
		ClaimableTokens: uint64(res.ClaimableTokens),
	}
	return grpcRes, nil
}

func EncodeGRPCRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(messages.RegisterResponse)
	grpcRes := &messages.RegisterResponse{
		Url: string(res.Url),
	}
	return grpcRes, nil
}
