package handlers

import (
	"context"
	"skadi_bot/utils"
	"strings"
	"sync"
	"time"

	pb "skadi_bot/proto"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	lastMessageVec []float32
	lastMessage    string
	msgLock        sync.Mutex
)

func CreateMsgHandler(client pb.Doc2VecServiceClient, db *utils.DB) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		timer := time.Now()
		defer func() {
			utils.TotalLatency.Observe(float64(time.Since(timer).Milliseconds()))
		}()

		msg := ctx.Event.Message.ExtractPlainText()
		utils.MessageRecCounter.Inc()

		utils.SLogger.Info("Received message", "uid", ctx.Event.UserID, "msg", msg, "source", "msg_handler")

		if msg[0] == '>' {
			msg = strings.Trim(msg, "> ")
		}

		if msg != "" && utils.ProbGeneratorManager.Get(utils.ProbTypeAIFeed) {
			utils.SLogger.Info("Feed message to AI", "msg", msg, "source", "msg_handler")
			utils.AIChatterClient.Feed(msg)
		}

		enableDoc2vec := utils.GetConfig().Doc2VecConfig.Enable
		var vec []float32
		if enableDoc2vec {

			doc2vecTimer := time.Now()
			resp, err := client.GetDoc2Vec(context.Background(), &pb.Doc2VecRequest{Text: msg})
			utils.Doc2vecLatency.Observe(float64(time.Since(doc2vecTimer).Milliseconds()))

			if err != nil {
				utils.SLogger.Warn("Failed to get vector", "msg", msg, "err", err, "source", "msg_handler")
				return
			}

			vec = resp.GetVector()

			msgLock.Lock()
			defer msgLock.Unlock()

			if lastMessage == "" || toBeFiltered(msg) {
				lastMessage = msg
				lastMessageVec = vec
				return
			}
			go db.SaveMessage(lastMessage, lastMessageVec, msg)
			lastMessage = msg
			lastMessageVec = vec
		}
		if msg != "" && utils.ProbGeneratorManager.Get(utils.ProbTypeAIResponse) {
			aiResp, err := utils.AIChatterClient.GetRespond(context.Background(), msg)
			if err != nil {
				utils.SLogger.Warn("Failed to get response from AI", "msg", msg, "err", err, "source", "msg_handler")
				return
			}
			if aiResp != "" {
				msgSend := "> " + aiResp
				utils.SLogger.Info("Send response", "msg", msg, "response", msgSend, "source", "msg_handler")
				utils.SendMsgCounter.Add(1)
				ctx.Send(msgSend)
				ctx.Block()
				return
			}
		}

		if enableDoc2vec {

			exists, next, err := db.MessageExists(msg)
			if exists {
				utils.MessageHitCounter.Inc()
			} else {
				utils.MessageMissCounter.Inc()
			}
			if err != nil {
				utils.SLogger.Warn("Failed to check message exists", "msg", msg, "err", err, "source", "msg_handler")
				return
			}
			utils.SLogger.Info("Message existence check", "msg", msg, "exists", exists, "next", next, "source", "msg_handler")
			if exists && utils.ProbGeneratorManager.Get(utils.ProbTypeHit) {
				utils.MessageHitReplyCounter.Inc()
				utils.SLogger.Info("Send response", "msg", msg, "next", next, "source", "msg_handler")
				utils.SendMsgCounter.Add(1)
				ctx.Send(next)
				ctx.Block()
				return
			}

			// not exist, need to find nearest message with vector
			nearestNext, err := db.GetNearestMessage(vec)
			if err != nil {
				utils.SLogger.Warn("Failed to get nearest message", "msg", msg, "err", err, "source", "msg_handler")
				return
			}
			if nearestNext != "" && utils.ProbGeneratorManager.Get(utils.ProbTypeMiss) {
				utils.MessageMissReplyCounter.Inc()
				utils.SLogger.Info("Send response", "msg", msg, "nearestNext", nearestNext, "source", "msg_handler")
				utils.SendMsgCounter.Add(1)
				ctx.Send(nearestNext)
				ctx.Block()
				return
			}
		}
	}
}

// If message is blank or contains url
func toBeFiltered(m string) bool {
	return m == "" || strings.Contains(m, "bilibili.com")
}
