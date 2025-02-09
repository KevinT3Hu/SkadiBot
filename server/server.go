package server

import "github.com/gin-gonic/gin"

type SkadiServer struct {
	r *gin.Engine
}

func NewSkadiServer() *SkadiServer {
	server := &SkadiServer{
		r: gin.Default(),
	}
	server.route()
	return server
}

func (s *SkadiServer) route() {
	s.routeUpdateConfig()
	s.r.GET("/config", GetConfig)
	s.r.GET("/health", HealthCheck)
}

func (s *SkadiServer) routeInfo() {
	g := s.r.Group("/info")
	g.GET("/uptime", GetUptime)
	g.GET("/msgCounter", GetMsgCounter)
}

func (s *SkadiServer) routeUpdateConfig() {
	g := s.r.Group("/update")
	g.Use(AuthMiddleware())
	g.POST("/doc2vecDestination", UpdateDoc2VecGrpcDestination)
	g.POST("/prob", UpdateProb)
	g.POST("/prompt", UpdatePrompt)
	g.POST("/aiBaseUrl", UpdateAIBaseUrl)
	g.POST("/aiKey", UpdateAIKey)
}

func (s *SkadiServer) Run() error {
	return s.r.Run(":8080")
}
