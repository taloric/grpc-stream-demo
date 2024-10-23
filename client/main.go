package main

import (
	"bufio"
	"client/message/pb"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ServerDelay = 10

func main() {
	port := "38888"
	host := "localhost"
	flag.StringVar(&port, "port", port, "The server port")
	flag.StringVar(&host, "host", host, "The server host")
	flag.IntVar(&ServerDelay, "delay", ServerDelay, "The server delay, unit: ms")
	flag.Parse()
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewStreamingServiceClient(conn)

	streamingClient := &StreamingClient{
		recvChan:   make(chan string),
		recvClient: client,
	}
	streamingClient.initOneClientStream()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Select the communication mode:")
		fmt.Println("1. Unary RPC")
		fmt.Println("2. Client Stream RPC")
		fmt.Println("3. Server Stream RPC")
		fmt.Println("4. Bidirectional Stream RPC")
		fmt.Println("5. Client Repeated Stream RPC, Use Same Stream during Send and Recv")
		fmt.Println("6. Exit")

		fmt.Print("Enter your choice (1-5): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			streamingClient.unaryRPC(client)
		case "2":
			streamingClient.clientStreamRPC(client)
		case "3":
			streamingClient.serverStreamRPC(client)
		case "4":
			streamingClient.bidirectionalStreamRPC(client)
		case "5":
			streamingClient.clientRepeatedStream()
		case "6":
			fmt.Println("Exiting...")
			streamingClient.recvChan <- "exit"
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

type StreamingClient struct {
	recvChan   chan string
	recvClient pb.StreamingServiceClient
}

func (s *StreamingClient) initOneClientStream() {
	go func() {
	streamLoop:
		for {
			stream, err := s.recvClient.ClientStreamRPC(context.Background())
			if err != nil {
				fmt.Printf("failed to call ClientStreamRPC: %v", err)
				continue streamLoop
			}
			for s := range s.recvChan {
				if s == "exit" {
					// not a good impl but for exit demo
					fmt.Println("receive exit")
					break
				}
				if err := stream.Send(&pb.ClientStreamRequest{Message: s}); err != nil {
					fmt.Printf("failed to send ClientStreamRequest: %v", err)
					continue streamLoop
				}
				fmt.Printf("Sending message: %s\n", s)
			}
			stream.CloseAndRecv()
			break
		}
	}()
}

func (s *StreamingClient) unaryRPC(client pb.StreamingServiceClient) {
	// Unary RPC
	unaryResponse, err := client.UnaryRPC(context.Background(), &pb.UnaryRequest{Message: "Hello, Unary RPC!"})
	if err != nil {
		log.Fatalf("failed to call UnaryRPC: %v", err)
	}
	fmt.Println(unaryResponse.GetResponse())
}

func (s *StreamingClient) clientStreamRPC(client pb.StreamingServiceClient) {
	// Client Stream RPC
	clientStream, err := client.ClientStreamRPC(context.Background())
	if err != nil {
		log.Fatalf("failed to call ClientStreamRPC: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := clientStream.Send(&pb.ClientStreamRequest{Message: fmt.Sprintf("Client Stream Message %d", i)}); err != nil {
			log.Fatalf("failed to send ClientStreamRequest: %v", err)
		}
		time.Sleep(time.Duration(ServerDelay) * time.Millisecond)
	}
	clientStreamResponse, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive ClientStreamResponse: %v", err)
	}
	fmt.Println(clientStreamResponse.GetResponse())
}

func (s *StreamingClient) clientRepeatedStream() {
	s.recvChan <- "Hello, 1"
	time.Sleep(100 * time.Millisecond)
	s.recvChan <- "Hello, 2"
	time.Sleep(100 * time.Millisecond)
	s.recvChan <- "Hello, 3"
}

func (s *StreamingClient) serverStreamRPC(client pb.StreamingServiceClient) {
	// Server Stream RPC
	serverStream, err := client.ServerStreamRPC(context.Background(), &pb.ServerStreamRequest{Message: "Hello, Server Stream RPC!"})
	if err != nil {
		log.Fatalf("failed to call ServerStreamRPC: %v", err)
	}
	for {
		resp, err := serverStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to receive ServerStreamResponse: %v", err)
		}
		fmt.Println(resp.GetResponse())
	}
}

func (s *StreamingClient) bidirectionalStreamRPC(client pb.StreamingServiceClient) {

	fmt.Println("Starting Bidirectional Stream RPC...")
	// Bidirectional Stream RPC
	bidirectionalStream, err := client.BidirectionalStreamRPC(context.Background())
	if err != nil {
		log.Fatalf("failed to call BidirectionalStreamRPC: %v", err)
	}
	go func() {
		for i := 0; i < 3; i++ {
			if err := bidirectionalStream.Send(&pb.BidirectionalStreamRequest{Message: fmt.Sprintf("Bidirectional Stream Message %d", i)}); err != nil {
				log.Fatalf("failed to send BidirectionalStreamRequest: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
		if err := bidirectionalStream.CloseSend(); err != nil {
			log.Fatalf("failed to close send stream: %v", err)
		}
	}()
	for {
		resp, err := bidirectionalStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to receive BidirectionalStreamResponse: %v", err)
		}
		fmt.Println(resp.GetResponse())
	}

}
