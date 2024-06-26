package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"server/message/pb"
	"time"

	"google.golang.org/grpc"
)

var ServerDelay = 10

type StreamingServer struct {
	pb.UnimplementedStreamingServiceServer
}

func (s *StreamingServer) UnaryRPC(ctx context.Context, req *pb.UnaryRequest) (*pb.UnaryResponse, error) {
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
	fmt.Printf("Server started, listening on %s \n", port)
	if err := s.Serve(l); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}

func main() {
	port := "38888"
	flag.StringVar(&port, "port", port, "The server port")
	flag.IntVar(&ServerDelay, "delay", ServerDelay, "The server delay, unit: ms")
	flag.Parse()
	server_start(port)
}
