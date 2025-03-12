package auth

import (
	"context"
	"net"
	"syscall"

	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"
	"p2p_remote_desk/server/config"

	"google.golang.org/grpc"
)

func Start() {
	serverConfig := config.GetConfig().Server
	listener, err := net.Listen("tcp", lkit.GetAddr(serverConfig.Host, serverConfig.AuthPort))
	if err != nil {
		llog.Error("Auth Server ListenTCP error:", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}

	// file, _ := credentials.NewServerTLSFromFile("", "")
	// grpc.Creds(file)

	ser := grpc.NewServer()
	registerServiceServer(ser)

	err = ser.Serve(listener)
	if err != nil {
		llog.Error("failed to serve: ", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}
}

func registerServiceServer(server *grpc.Server) {
	common.RegisterAuthServer(server, &Login{})
}

type Login struct {
	common.UnimplementedAuthServer
}

func (s *Login) Login(_ context.Context, req *common.LoginRequest) (*common.LoginResponse, error) {
	if req.GetUsername() != "admin" {
		return &common.LoginResponse{
			Code: common.AuthCode_InvalidUsername,
		}, nil
	}
	if req.GetPassword() != "222" {
		return &common.LoginResponse{
			Code: common.AuthCode_InvalidPassword,
		}, nil
	}
	return &common.LoginResponse{
		Code: common.AuthCode_OK,
	}, nil
}
