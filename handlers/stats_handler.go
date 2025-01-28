package handlers

import (
	"fmt"
	"runtime"
	"skadi_bot/utils"
	"syscall"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func CreateStatsHandler(db *utils.DB, aiChatter *utils.AIChatter) func(*zero.Ctx) {
	return func(ctx *zero.Ctx) {
		ctx.Block()
		utils.SLogger.Info("Getting stats", "source", "stats_handler")
		// get memory usage
		pUsage, total, free, err := getMemoryUsage()
		if err != nil {
			ctx.Send("Failed to get memory usage")
			utils.SLogger.Warn("Failed to get memory usage", "err", err, "source", "stats_handler")
			return
		}

		// get db size
		dbSize, tableSize, err := db.GetSize()
		if err != nil {
			ctx.Send("Failed to get db size")
			utils.SLogger.Warn("Failed to get db size", "err", err, "source", "stats_handler")
			return
		}

		dbRowCount, err := db.GetRowCount()
		if err != nil {
			ctx.Send("Failed to get row count")
			utils.SLogger.Warn("Failed to get row count", "err", err, "source", "stats_handler")
		}

		upTime := timeToHumanReadable(time.Since(utils.StartTime))

		contextLength := aiChatter.GetChatContextLength()

		msg := fmt.Sprintf("Memory usage: %s\nTotal: %s\nFree: %s\nDB size: %s\nTable size: %s\nDB row count: %d\nUp time: %s\nChat Context Length: %d", pUsage, total, free, dbSize, tableSize, dbRowCount, upTime, contextLength)

		ctx.Send(msg)
	}
}

func timeToHumanReadable(d time.Duration) string {
	days := d / (time.Hour * 24)
	d -= days * (time.Hour * 24)
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	return fmt.Sprintf("%d days, %d hours, %d minutes, %d seconds", days, hours, minutes, seconds)
}

func getMemoryUsage() (string, string, string, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// get memory usage of this program
	pUsage := bToHumanReadable(m.Sys)

	sysInfo := syscall.Sysinfo_t{}
	err := syscall.Sysinfo(&sysInfo)
	if err != nil {
		return "", "", "", err
	}
	// get total
	total := bToHumanReadable(uint64(sysInfo.Totalram) * uint64(sysInfo.Unit))
	// get free
	free := bToHumanReadable(uint64(sysInfo.Freeram) * uint64(sysInfo.Unit))

	return pUsage, total, free, nil
}

func bToHumanReadable(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
