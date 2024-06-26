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

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Select the communication mode:")
		fmt.Println("1. Unary RPC")
		fmt.Println("2. Client Stream RPC")
		fmt.Println("3. Server Stream RPC")
		fmt.Println("4. Bidirectional Stream RPC")
		fmt.Println("5. Exit")

		fmt.Print("Enter your choice (1-5): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			unaryRPC(client)
		case "2":
			clientStreamRPC(client)
		case "3":
			serverStreamRPC(client)
		case "4":
			bidirectionalStreamRPC(client)
		case "5":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func unaryRPC(client pb.StreamingServiceClient) {
	// Unary RPC
	unaryResponse, err := client.UnaryRPC(context.Background(), &pb.UnaryRequest{Message: "Hello, Unary RPC!"})
	if err != nil {
		log.Fatalf("failed to call UnaryRPC: %v", err)
	}
	fmt.Println(unaryResponse.GetResponse())
}

func clientStreamRPC(client pb.StreamingServiceClient) {
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

func serverStreamRPC(client pb.StreamingServiceClient) {
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

func bidirectionalStreamRPC(client pb.StreamingServiceClient) {

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
