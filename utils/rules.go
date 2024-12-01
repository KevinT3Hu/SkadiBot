package utils

import zero "github.com/wdvxdr1123/ZeroBot"

func NewGroupCheckRule(groupId int64) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		return ctx.Event.GroupID == groupId
	}
}

func NewAtMeRule() func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		return ctx.Event.IsToMe
	}
}

func NewIsAdminRule(admin int64) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		return ctx.Event.Sender.ID == admin
	}
}
