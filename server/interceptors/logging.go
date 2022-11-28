package interceptors

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// LoggingInterceptor is a server interceptor that logs the request and response
// of each RPC.
func UnaryLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p, _ := peer.FromContext(ctx)
	ip, _, _ := net.SplitHostPort(p.Addr.String())
	
	log.Printf("%v request from %v", info.FullMethod, ip) 

	resp, err := handler(ctx, req)
	return resp, err
}
	