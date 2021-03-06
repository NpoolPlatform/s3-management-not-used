package main

import (
	"time"

	"github.com/NpoolPlatform/s3-management/api"
	msgcli "github.com/NpoolPlatform/s3-management/pkg/message/client"
	msglistener "github.com/NpoolPlatform/s3-management/pkg/message/listener"
	msg "github.com/NpoolPlatform/s3-management/pkg/message/message"
	msgsrv "github.com/NpoolPlatform/s3-management/pkg/message/server"
	"golang.org/x/xerrors"

	grpc2 "github.com/NpoolPlatform/go-service-framework/pkg/grpc"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/oss"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	ossconst "github.com/NpoolPlatform/go-service-framework/pkg/oss/const"

	cli "github.com/urfave/cli/v2"

	"google.golang.org/grpc"
)

const BucketKey = "kyc-bucket"

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"s"},
	Usage:   "Run the daemon",
	Action: func(c *cli.Context) error {
		err := oss.Init(ossconst.SecretStoreKey, BucketKey)
		if err != nil {
			return xerrors.Errorf("fail to init s3: %v", err)
		}

		go func() {
			if err := grpc2.RunGRPC(rpcRegister); err != nil {
				logger.Sugar().Errorf("fail to run grpc server: %v", err)
			}
		}()

		if err := msgsrv.Init(); err != nil {
			return err
		}
		if err := msgcli.Init(); err != nil {
			return err
		}

		go msglistener.Listen()
		go msgSender()

		return grpc2.RunGRPCGateWay(rpcGatewayRegister)
	},
}

func rpcRegister(server grpc.ServiceRegistrar) error {
	api.Register(server)
	return nil
}

func rpcGatewayRegister(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return api.RegisterGateway(mux, endpoint, opts)
}

func msgSender() {
	id := 0
	for {
		logger.Sugar().Infof("send example")
		err := msgsrv.PublishExample(&msg.Example{
			ID:      id,
			Example: "hello world",
		})
		if err != nil {
			logger.Sugar().Errorf("fail to send example: %v", err)
			return
		}
		id++
		time.Sleep(3 * time.Second)
	}
}
