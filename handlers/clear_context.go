package handlers

import (
	"skadi_bot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func CreateClearContextHandler(aiChatter *utils.AIChatter) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		utils.SLogger.Info("Clearing chat context", "source", "clear_context")
		ctx.Block()
		aiChatter.ClearChatContext()
		utils.SLogger.Info("Chat context cleared", "source", "clear_context")
		ctx.Send("Chat context cleared")
	}
}
