package handlers

import (
	"skadi_bot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func CreateRecvMsgHandler() func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		utils.RecvMsgCounter.Add(1)
	}
}
