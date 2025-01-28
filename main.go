package main

import (
	"os"
	"skadi_bot/handlers"
	pb "skadi_bot/proto"
	"skadi_bot/utils"
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	utils.SLogger.Info("Starting bot")
	grpc_addr := os.Getenv("GRPC_ADDR")
	grpc_port := os.Getenv("GRPC_PORT")
	if grpc_addr == "" || grpc_port == "" {
		utils.SLogger.Error("GRPC_ADDR or GRPC_PORT is empty", "grpc_addr", grpc_addr, "grpc_port", grpc_port)
		os.Exit(1)
	}
	addr := grpc_addr + ":" + grpc_port
	utils.SLogger.Info("Get GRPC address", "addr", addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.SLogger.Error("Failed to create GRPC client", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewDoc2VecServiceClient(conn)

	aiChatter, err := utils.NewAiChatter()
	if err != nil {
		utils.SLogger.Error("Failed to create AIChatter", "err", err)
		os.Exit(1)
	}

	db, err := utils.NewDB()
	if err != nil {
		utils.SLogger.Error("Failed to create DB", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	groupId, err := strconv.ParseInt(os.Getenv("GROUP_ID"), 10, 64)
	if err != nil {
		utils.SLogger.Error("Failed to parse GROUP_ID", "err", err)
		os.Exit(1)
	}
	adminId, err := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	if err != nil {
		utils.SLogger.Error("Failed to parse ADMIN_ID", "err", err)
		os.Exit(1)
	}

	go utils.StartMetric()

	zero.OnCommand("$ct", utils.NewGroupCheckRule(groupId), utils.NewIsAdminRule(adminId)).Handle(handlers.CreateClearContextHandler(aiChatter))
	zero.OnCommand("$stats", utils.NewGroupCheckRule(groupId)).Handle(handlers.CreateStatsHandler(db, aiChatter))
	zero.OnCommand("$rv", utils.NewIsAdminRule(adminId)).Handle(handlers.CreateRebuildHandler(client, db))
	zero.OnMessage(utils.NewGroupCheckRule(groupId), utils.NewAtMeRule()).Handle(handlers.CreateAtMeHandler(aiChatter))
	zero.OnMessage(utils.NewGroupCheckRule(groupId)).Handle(handlers.CreateMsgHandler(client, aiChatter, db))

	utils.SLogger.Info("Start Run")
	wsAddr := os.Getenv("WS_ADDR")
	if wsAddr == "" {
		utils.SLogger.Error("WS_ADDR is empty")
		os.Exit(1)
	}
	zero.RunAndBlock(&zero.Config{
		Driver: []zero.Driver{
			driver.NewWebSocketClient(wsAddr, ""),
		},
	}, nil)
}
