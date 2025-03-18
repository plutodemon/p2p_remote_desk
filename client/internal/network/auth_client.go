package network

import (
	"context"
	"time"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func LoginAuth(userName, passWord string) int32 {
	// _, _ = credentials.NewClientTLSFromFile("server.crt", "server.grpc.io")
	cfg := config.GetConfig().ServerConfig
	conn, err := grpc.NewClient(lkit.GetAddr(cfg.Address, cfg.AuthPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		llog.Error("Did not connect: ", err)
		return -1
	}
	defer conn.Close()

	client := common.NewAuthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rsp, err := client.Login(ctx, &common.LoginRequest{
		Username: userName,
		Password: passWord,
	})
	if err != nil {
		llog.Error("could not greet: ", err)
		return -1
	}

	authCode := rsp.GetCode()
	if authCode != common.AuthCode_OK {
		llog.Error("login failed: ", authCode)
	}

	return int32(authCode)
}
