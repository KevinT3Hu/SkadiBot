package handlers

import (
	"context"
	"skadi_bot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func CreateAtMeHandler(aiChatter *utils.AIChatter) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		// Get the message content
		msg := ctx.ExtractPlainText()
		utils.SLogger.Info("Received AT message", "uid", ctx.Event.UserID, "msg", msg, "source", "at_me")

		// Call the AI to get the response
		response, err := aiChatter.GetAtRespond(context.Background(), msg)
		if err != nil {
			utils.SLogger.Warn("Failed to get response from AI", "err", err, "source", "at_me")
			return
		}
		ctx.Block()

		// Send the response
		utils.SLogger.Info("Send response", "response", response, "source", "at_me")
		ctx.Send(response)
	}
}
