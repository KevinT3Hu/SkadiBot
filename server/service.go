package server

import (
	"skadi_bot/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UpdateDoc2VecGrpcDestination(c *gin.Context) {
	newDestination := c.PostForm("destination")
	if newDestination == "" {
		utils.SLogger.Warn("Get empty destination", "source", "server")
		c.JSON(400, gin.H{"error": "destination is required"})
		return
	}
	// Update the destination
	utils.SLogger.Info("UpdateDoc2VecGrpcDestination", "destination", newDestination)
	utils.Doc2VecAddrChan <- newDestination
	c.JSON(200, gin.H{"message": "destination updated"})
}

func UpdateProb(c *gin.Context) {
	probTypeString := c.PostForm("probType")
	probType := utils.ProbType(probTypeString)
	newProb := c.PostForm("prob")
	if newProb == "" {
		utils.SLogger.Warn("Get empty prob", "source", "server")
		c.JSON(400, gin.H{"error": "prob is required"})
		return
	}
	prob, err := strconv.ParseFloat(newProb, 64)
	if err != nil {
		utils.SLogger.Warn("Get invalid prob", "prob", newProb, "source", "server")
		c.JSON(400, gin.H{"error": "prob must be a float"})
		return
	}
	// Update the prob
	utils.SLogger.Info("UpdateProb", "probType", probType, "prob", prob, "source", "server")
	utils.ProbGeneratorManager.UpdateProb(probType, prob)
	c.JSON(200, gin.H{"message": "prob updated"})
}

func UpdatePromptPreset(c *gin.Context) {
	presetName := c.PostForm("name")
	if presetName == "" {
		utils.SLogger.Warn("Get empty preset name", "source", "server")
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}
	systemPrompt := c.PostForm("systemPrompt")
	atPrompt := c.PostForm("atPrompt")
	config := utils.GetConfig()
	preset := utils.PromptPreset{
		AIRequestPrompt:   systemPrompt,
		AIAtRequestPrompt: atPrompt,
	}
	config.PromptConfig.Presets[presetName] = preset
	utils.UpdateConfig(config)
	c.JSON(200, gin.H{"message": "preset updated"})
}

func UpdateAIBaseUrl(c *gin.Context) {
	newBaseUrl := c.PostForm("baseUrl")
	if newBaseUrl == "" {
		utils.SLogger.Warn("Get empty baseUrl", "source", "server")
		c.JSON(400, gin.H{"error": "baseUrl is required"})
		return
	}
	// Update the baseUrl
	utils.SLogger.Info("UpdateAIBaseUrl", "baseUrl", newBaseUrl, "source", "server")
	config := utils.GetConfig()
	config.AIApiConfig.BaseUrl = newBaseUrl
	utils.UpdateConfig(config)
	utils.AIChatterClient.NewConfig(utils.GetConfig().AIApiConfig)
	c.JSON(200, gin.H{"message": "baseUrl updated"})
}

func UpdateAIKey(c *gin.Context) {
	newKey := c.PostForm("key")
	if newKey == "" {
		utils.SLogger.Warn("Get empty key", "source", "server")
		c.JSON(400, gin.H{"error": "key is required"})
		return
	}
	// Update the key
	utils.SLogger.Info("UpdateAIKey", "key", newKey, "source", "server")
	config := utils.GetConfig()
	config.AIApiConfig.APIKey = newKey
	utils.UpdateConfig(config)
	utils.AIChatterClient.NewConfig(utils.GetConfig().AIApiConfig)
	c.JSON(200, gin.H{"message": "key updated"})
}

func GetConfig(c *gin.Context) {
	config := utils.GetConfig()
	// redact the key
	config.AIApiConfig.APIKey = "REDACTED"
	c.JSON(200, gin.H{"config": config})
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"message": "ok"})
}

func GetUptime(c *gin.Context) {
	startTime := utils.StartTime
	uptime := time.Now().Sub(startTime)
	c.JSON(200, gin.H{"uptime": uptime.Seconds()})
}

func GetMsgCounter(c *gin.Context) {
	recvMsgCounter := utils.RecvMsgCounter.Load()
	sendMsgCounter := utils.SendMsgCounter.Load()
	c.JSON(200, gin.H{"recvMsgCounter": recvMsgCounter, "sendMsgCounter": sendMsgCounter})
}
