package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type HelloService struct {
	UnimplementedHelloServiceServer
}

func (p *HelloService) Hello(ctx context.Context, in *String) (*String, error) {
	return &String{Value: "hello" + in.GetValue()}, nil
}

func Test_RpcServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	// file, _ := credentials.NewServerTLSFromFile("", "")
	// grpc.Creds(file)
	s := grpc.NewServer()
	RegisterHelloServiceServer(s, &HelloService{})
	// 启动服务
	err = s.Serve(listener)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}

func Test_RpcClient(t *testing.T) {
	_, _ = credentials.NewClientTLSFromFile("server.crt", "server.grpc.io")
	conn, err := grpc.NewClient("localhost:1234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewHelloServiceClient(conn)

	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Hello(ctx, &String{Value: "客户端"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Value)
}
