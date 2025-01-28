package handlers

import (
	"context"
	"skadi_bot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"go.uber.org/zap"
)

func CreateAtMeHandler(suger *zap.SugaredLogger, aiChatter *utils.AIChatter) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		// Get the message content
		msg := ctx.ExtractPlainText()
		suger.Infof("Received AT message: %s", msg)

		// Call the AI to get the response
		response, err := aiChatter.GetAtRespond(context.Background(), msg)
		if err != nil {
			suger.Errorf("Failed to get response from AI: %v", err)
			return
		}
		ctx.Block()

		// Send the response
		ctx.Send(response)
	}
}
