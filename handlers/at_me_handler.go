package handlers

import (
	"context"
	pb "skadi_bot/proto"
	"skadi_bot/utils"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"go.uber.org/zap"
)

func CreateAtMeHandler(suger *zap.SugaredLogger, client pb.Doc2VecServiceClient, db *utils.DB) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		timer := time.Now()
		defer func() {
			utils.TotalLatency.Observe(float64(time.Since(timer).Milliseconds()))
		}()

		ctx.Block()

		msg := ctx.Event.Message.ExtractPlainText()
		utils.MessageRecCounter.Inc()
		doc2vecTimer := time.Now()
		resp, err := client.GetDoc2Vec(context.Background(), &pb.Doc2VecRequest{Text: msg})
		utils.Doc2vecLatency.Observe(float64(time.Since(doc2vecTimer).Milliseconds()))
		if err != nil {
			suger.Errorf("Failed to get vector: %v", err)
			return
		}
		vec := resp.GetVector()
		exists, next, err := db.MessageExists(msg)
		if err != nil {
			suger.Errorf("Failed to check message exists: %v", err)
			return
		}
		suger.Infow("Message exists", "exists", exists, "next", next)
		if exists {
			suger.Infow("Send next message", "next", next)
			ctx.Send(next)
			return
		}
		// not exist, need to find nearest message with vector
		nearestNext, err := db.GetNearestMessage(vec)
		if err != nil {
			suger.Errorf("Failed to get nearest message: %v", err)
			return
		}
		if nearestNext != "" {
			suger.Infow("Send nearest message", "nearestNext", nearestNext)
			ctx.Send(nearestNext)
			return
		}
	}
}
