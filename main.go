package main

import (
	"log"
	"os"
	"skadi_bot/handlers"
	pb "skadi_bot/proto"
	"skadi_bot/utils"
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Printf("Failed to create logger: %v, no logs will be printed thus far", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Info("Starting bot")
	grpc_addr := os.Getenv("GRPC_ADDR")
	grpc_port := os.Getenv("GRPC_PORT")
	if grpc_addr == "" || grpc_port == "" {
		sugar.Fatal("GRPC_ADDR or GRPC_PORT is empty")
	}
	addr := grpc_addr + ":" + grpc_port
	log.Printf("GRPC_ADDR: %s", addr)
	sugar.Infow("Get GRPC_ADDR", "addr", addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		sugar.Fatal("Failed to create GRPC client: %v", err)
	}
	defer conn.Close()

	client := pb.NewDoc2VecServiceClient(conn)

	db, err := utils.NewDB()
	if err != nil {
		sugar.Fatal("Failed to create DB: %v", err)
	}
	defer db.Close()

	groupId, err := strconv.ParseInt(os.Getenv("GROUP_ID"), 10, 64)
	if err != nil {
		sugar.Fatalf("Failed to parse GROUP_ID: %v", err)
	}
	adminId, err := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	if err != nil {
		sugar.Fatalf("Failed to parse ADMIN_ID: %v", err)
	}

	go utils.StartMetric()

	zero.OnCommand("$stats", utils.NewGroupCheckRule(groupId)).Handle(handlers.CreateStatsHandler(sugar, db))
	zero.OnCommand("$rv", utils.NewIsAdminRule(adminId)).Handle(handlers.CreateRebuildHandler(sugar, client, db))
	zero.OnMessage(utils.NewGroupCheckRule(groupId), utils.NewAtMeRule()).Handle(handlers.CreateAtMeHandler(sugar, client, db))
	zero.OnMessage(utils.NewGroupCheckRule(groupId)).Handle(handlers.CreateMsgHandler(sugar, client, db))

	sugar.Infof("Start Run")
	wsAddr := os.Getenv("WS_ADDR")
	if wsAddr == "" {
		sugar.Fatal("WS_ADDR is empty")
	}
	zero.RunAndBlock(&zero.Config{
		Driver: []zero.Driver{
			driver.NewWebSocketClient(wsAddr, ""),
		},
	}, nil)
}
