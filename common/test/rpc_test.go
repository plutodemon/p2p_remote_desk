package test

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
)

type HelloServiceInterface interface {
	Hello(request *String, reply *string) error
}

type HelloService struct{}

func (p *HelloService) Hello(request *String, reply *string) error {
	*reply = "hello:" + request.GetValue()
	return nil
}

const HelloServiceName = "HelloService"

func RegisterHelloService(svc HelloServiceInterface) error {
	return rpc.RegisterName(HelloServiceName, svc)
}

func Test_RpcServer(t *testing.T) {
	err := RegisterHelloService(new(HelloService))
	if err != nil {
		return
	}

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Accept error:", err)
	}

	rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
}

func Test_RpcClient(t *testing.T) {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	// client := rpc.NewClient(conn)
	var reply string
	err = client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		log.Fatal("Call error:", err)
	}
	log.Println(reply)
}
