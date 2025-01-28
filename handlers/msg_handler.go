package handlers

import (
	"context"
	"os"
	"skadi_bot/utils"
	"strconv"
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

func CreateMsgHandler(client pb.Doc2VecServiceClient, aiChatter *utils.AIChatter, db *utils.DB) func(ctx *zero.Ctx) {
	nonHitProb, err := strconv.ParseFloat(os.Getenv("NON_HIT_PROB"), 10)
	if err != nil {
		nonHitProb = 0.05
	}
	utils.SLogger.Info("nonHitProb", "nonHitProb", nonHitProb, "souece", "CreateMsgHandler")

	hitProb, err := strconv.ParseFloat(os.Getenv("HIT_PROB"), 10)
	if err != nil {
		hitProb = 0.1
	}
	utils.SLogger.Info("hitProb", "hitProb", hitProb, "souece", "CreateMsgHandler")

	aiFeedProb, err := strconv.ParseFloat(os.Getenv("AI_FEED_PROB"), 10)
	if err != nil {
		aiFeedProb = 0.8
	}
	utils.SLogger.Info("aiFeedProb", "aiFeedProb", aiFeedProb, "souece", "CreateMsgHandler")

	aiRequestProb, err := strconv.ParseFloat(os.Getenv("AI_REQUEST_PROB"), 10)
	if err != nil {
		aiRequestProb = 0.1
	}
	utils.SLogger.Info("aiRequestProb", "aiRequestProb", aiRequestProb, "souece", "CreateMsgHandler")

	hitPb := utils.NewProbGenerator(hitProb)
	nonHitPb := utils.NewProbGenerator(nonHitProb)
	aiFeedPb := utils.NewProbGenerator(aiFeedProb)
	aiReqPb := utils.NewProbGenerator(aiRequestProb)

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

		if msg != "" && aiFeedPb.Get() {
			utils.SLogger.Info("Feed message to AI", "msg", msg, "source", "msg_handler")
			aiChatter.Feed(msg)
		}

		doc2vecTimer := time.Now()
		resp, err := client.GetDoc2Vec(context.Background(), &pb.Doc2VecRequest{Text: msg})
		utils.Doc2vecLatency.Observe(float64(time.Since(doc2vecTimer).Milliseconds()))

		if err != nil {
			utils.SLogger.Warn("Failed to get vector", "msg", msg, "err", err, "source", "msg_handler")
			return
		}

		vec := resp.GetVector()

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

		if msg != "" && aiReqPb.Get() {

			aiResp, err := aiChatter.GetRespond(context.Background(), msg)
			if err != nil {
				utils.SLogger.Warn("Failed to get response from AI", "msg", msg, "err", err, "source", "msg_handler")
				return
			}
			if aiResp != "" {
				msgSend := "> " + aiResp
				utils.SLogger.Info("Send response", "msg", msg, "response", msgSend, "source", "msg_handler")
				ctx.Send(msgSend)
				ctx.Block()
				return
			}
		}

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
		if exists && hitPb.Get() {
			utils.MessageHitReplyCounter.Inc()
			utils.SLogger.Info("Send response", "msg", msg, "next", next, "source", "msg_handler")
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
		if nearestNext != "" && nonHitPb.Get() {
			utils.MessageMissReplyCounter.Inc()
			utils.SLogger.Info("Send response", "msg", msg, "nearestNext", nearestNext, "source", "msg_handler")
			ctx.Send(nearestNext)
			ctx.Block()
			return
		}
	}
}

// If message is blank or contains url
func toBeFiltered(m string) bool {
	return m == "" || strings.Contains(m, "bilibili.com")
}
