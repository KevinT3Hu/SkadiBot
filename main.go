package main

import (
	"os"
	"skadi_bot/handlers"
	pb "skadi_bot/proto"
	"skadi_bot/server"
	"skadi_bot/utils"
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

func main() {
	utils.SLogger.Info("Starting bot")

	resolver.Register(&utils.Doc2VecGrpcResolverBuilder{})
	conn, err := grpc.NewClient("doc2vec:///1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.SLogger.Error("Failed to create GRPC client", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewDoc2VecServiceClient(conn)

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

	skadiServer := server.NewSkadiServer()
	go skadiServer.Run()

	zero.OnCommand("$ct", utils.NewGroupCheckRule(groupId), utils.NewIsAdminRule(adminId)).Handle(handlers.CreateClearContextHandler()).SetPriority(0)
	zero.OnCommand("$stats", utils.NewGroupCheckRule(groupId)).Handle(handlers.CreateStatsHandler(db)).SetPriority(1)
	zero.OnCommand("$rv", utils.NewIsAdminRule(adminId)).Handle(handlers.CreateRebuildHandler(client, db)).SetPriority(2)
	zero.OnMessage(utils.NewGroupCheckRule(groupId), utils.NewAtMeRule()).Handle(handlers.CreateAtMeHandler()).SetPriority(3)
	zero.OnMessage(utils.NewGroupCheckRule(groupId)).Handle(handlers.CreateMsgHandler(client, db)).SetPriority(4)

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
