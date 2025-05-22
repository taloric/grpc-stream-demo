package main

import (
	"context"
	"fmt"
	"io"
	"net"

	// "net/http"
	"server/message/pb"
	"time"

	// "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var ServerDelay = 10

type StreamingServer struct {
	pb.UnimplementedStreamingServiceServer
}

func displayMetadata(ctx context.Context) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Println("receive metadata: {")
		for k, v := range md {
			fmt.Printf("key: %s, val: %s \n", k, v)
		}
		fmt.Println("}")
	}
}

func (s *StreamingServer) UnaryRPC(ctx context.Context, req *pb.UnaryRequest) (*pb.UnaryResponse, error) {
	displayMetadata(ctx)
	return &pb.UnaryResponse{Response: "Unary RPC response: " + req.GetMessage()}, nil
}

func (s *StreamingServer) ClientStreamRPC(stream pb.StreamingService_ClientStreamRPCServer) error {
	var messages []string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.ClientStreamResponse{Response: fmt.Sprintf("Client Stream RPC response: %v", messages)})
		}
		if err != nil {
			return err
		}
		messages = append(messages, req.GetMessage())
	}
}

func (s *StreamingServer) ServerStreamRPC(req *pb.ServerStreamRequest, stream pb.StreamingService_ServerStreamRPCServer) error {
	for i := 0; i < 3; i++ {
		if err := stream.Send(&pb.ServerStreamResponse{Response: fmt.Sprintf("Server Stream RPC response %d: %s", i, req.GetMessage())}); err != nil {
			return err
		}
		time.Sleep(time.Duration(ServerDelay) * time.Millisecond)
	}
	return nil
}

func (s *StreamingServer) BidirectionalStreamRPC(stream pb.StreamingService_BidirectionalStreamRPCServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.BidirectionalStreamResponse{Response: "Bidirectional Stream RPC response: " + req.GetMessage()}); err != nil {
			return err
		}
		time.Sleep(time.Duration(ServerDelay) * time.Millisecond)
	}
}

func server_start(port string) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStreamingServiceServer(s, &StreamingServer{})

	// server mux for handle http&grpc req
	// mux := runtime.NewServeMux()
	// mux.HandlePath("GET", "/hello", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// 	fmt.Println("http request touched")
	// 	w.Write([]byte("hello, http"))
	// }))

	// go func() {
	// 	if err := http.Serve(l, mux); err != nil {
	// 		fmt.Printf("failed to serve http: %v", err)
	// 	}
	// }()

	if err := s.Serve(l); err != nil {
		fmt.Printf("failed to serve grpc: %v", err)
	}
}
