package handlers

import (
	"context"
	"os"
	"skadi_bot/utils"
	"strconv"
	"sync"
	"time"

	pb "skadi_bot/proto"

	zero "github.com/wdvxdr1123/ZeroBot"
	"go.uber.org/zap"
)

var (
	lastMessage string
	msgLock     sync.Mutex
)

func CreateMsgHandler(sugar *zap.SugaredLogger, client pb.Doc2VecServiceClient, db *utils.DB) func(ctx *zero.Ctx) {

	nonHitProb, err := strconv.ParseFloat(os.Getenv("NON_HIT_PROB"), 10)
	if err != nil {
		nonHitProb = 0.05
	}

	hitProb, err := strconv.ParseFloat(os.Getenv("HIT_PROB"), 10)
	if err != nil {
		hitProb = 0.1
	}

	hitPb := utils.NewProbGenerator(hitProb)
	nonHitPb := utils.NewProbGenerator(nonHitProb)

	return func(ctx *zero.Ctx) {
		timer := time.Now()
		defer func() {
			utils.TotalLatency.Observe(float64(time.Since(timer).Milliseconds()))
		}()

		msg := ctx.Event.Message.ExtractPlainText()
		utils.MessageRecCounter.Inc()

		doc2vecTimer := time.Now()
		resp, err := client.GetDoc2Vec(context.Background(), &pb.Doc2VecRequest{Text: msg})
		utils.Doc2vecLatency.Observe(float64(time.Since(doc2vecTimer).Milliseconds()))

		if err != nil {
			sugar.Errorf("Failed to get vector: %v", err)
			return
		}

		msgLock.Lock()
		defer msgLock.Unlock()

		if lastMessage == "" || msg == "" {
			lastMessage = msg
			return
		}
		vec := resp.GetVector()
		go db.SaveMessage(lastMessage, vec, msg)
		lastMessage = msg

		exists, next, err := db.MessageExists(msg)
		if exists {
			utils.MessageHitCounter.Inc()
		} else {
			utils.MessageMissCounter.Inc()
		}
		if err != nil {
			sugar.Errorf("Failed to check message exists: %v", err)
			return
		}
		sugar.Infow("Message exists", "exists", exists, "next", next)
		if exists && hitPb.Get() {
			utils.MessageHitReplyCounter.Inc()
			sugar.Infow("Send next message", "next", next)
			ctx.Send(next)
			return
		}

		// not exist, need to find nearest message with vector
		nearestNext, err := db.GetNearestMessage(vec)
		if err != nil {
			sugar.Errorf("Failed to get nearest message: %v", err)
			return
		}
		if nearestNext != "" && nonHitPb.Get() {
			utils.MessageMissReplyCounter.Inc()
			sugar.Infow("Send nearest message", "nearestNext", nearestNext)
			ctx.Send(nearestNext)
			return
		}
	}
}
