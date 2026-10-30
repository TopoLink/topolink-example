// Package grpcapi is the internal order lookup API.
package grpcapi

import (
	"context"

	"google.golang.org/grpc"

	"github.com/topolink/topolink-example/internal/service"
)

// GetOrderRequest is the lookup request.
type GetOrderRequest struct {
	OrderID int64 `json:"order_id"`
}

// GetOrderResponse is the lookup response.
type GetOrderResponse struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Total   int64  `json:"total"`
}

// OrderServiceServer is the service contract.
type OrderServiceServer interface {
	GetOrder(context.Context, *GetOrderRequest) (*GetOrderResponse, error)
}

// OrderServer implements OrderServiceServer.
type OrderServer struct {
	Orders *service.OrderService
}

// GetOrder returns one order for internal callers.
func (s *OrderServer) GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	o, err := s.Orders.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	return &GetOrderResponse{OrderID: o.ID, Status: string(o.Status), Total: o.Total}, nil
}

// RegisterOrderServiceServer registers the service. This is the hand-written
// equivalent of the generated registration; replace it with protoc output when
// order.proto is added.
func RegisterOrderServiceServer(s *grpc.Server, srv OrderServiceServer) {
	s.RegisterService(&orderServiceDesc, srv)
}

var orderServiceDesc = grpc.ServiceDesc{
	ServiceName: "order.v1.OrderService",
	HandlerType: (*OrderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetOrder",
			Handler: func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
				in := new(GetOrderRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				return srv.(OrderServiceServer).GetOrder(ctx, in)
			},
		},
	},
	Metadata: "order.proto",
}
