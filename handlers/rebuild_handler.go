package handlers

import (
	"context"
	pb "skadi_bot/proto"
	"skadi_bot/utils"
	"strconv"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"go.uber.org/zap"
)

var RebuildLock sync.Mutex

func CreateRebuildHandler(sugar *zap.SugaredLogger, client pb.Doc2VecServiceClient, db *utils.DB) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		ctx.Block()

		RebuildLock.Lock()
		defer RebuildLock.Unlock()

		ctx.Send("Rebuilding message vector")

		timer := time.Now()
		defer func() {
			elapsed := time.Since(timer).Milliseconds()
			msg := "Rebuild message vector took " + strconv.FormatInt(elapsed, 10) + "ms"
			ctx.Send(msg)
			sugar.Info(msg)
		}()
		// Rebuild the model
		err := db.RebuildMessageVec(func(s string) ([]float32, error) {
			resp, err := client.GetDoc2Vec(context.Background(), &pb.Doc2VecRequest{Text: s})
			if err != nil {
				return nil, err
			}
			return resp.Vector, nil
		})
		if err != nil {
			ctx.Send("Failed to rebuild message vector: " + err.Error())
			sugar.Errorf("Failed to rebuild message vector: %v", err)
			return
		}
	}
}
