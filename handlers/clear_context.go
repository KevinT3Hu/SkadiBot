package handlers

import (
	"skadi_bot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"go.uber.org/zap"
)

func CreateClearContextHandler(sugar *zap.SugaredLogger, aiChatter *utils.AIChatter) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		ctx.Block()
		aiChatter.ClearChatContext()
		sugar.Info("Chat context cleared")
		ctx.Send("Chat context cleared")
	}
}
