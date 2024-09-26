package server

import (
	"github.com/cynic-1/blockchain-teaching-system/internal/api"
	"github.com/cynic-1/blockchain-teaching-system/internal/auth"
	"github.com/cynic-1/blockchain-teaching-system/internal/config"
	"github.com/cynic-1/blockchain-teaching-system/internal/database"
	"github.com/cynic-1/blockchain-teaching-system/internal/docker"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	config *config.Config
	db     *database.Database
	docker *docker.DockerManager
}

func NewServer(config *config.Config) (*Server, error) {
	db, err := database.NewDatabase(config.BadgerDBPath)
	if err != nil {
		return nil, err
	}

	dockerManager, err := docker.NewDockerManager(config.DockerAPIVersion)
	if err != nil {
		return nil, err
	}
	//
	//err = auth.InitSecretKey()
	//if err != nil {
	//	return nil, err
	//}

	server := &Server{
		router: gin.Default(),
		config: config,
		db:     db,
		docker: dockerManager,
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() {
	handler := &api.Handler{
		DB:     s.db,
		Docker: s.docker,
	}
	// 公开路由组，不需要 token 验证
	public := s.router.Group("/api")
	{
		public.POST("/register", handler.Register)
		public.POST("/login", handler.Login)
	}

	// 受保护的路由组，需要 token 验证
	protected := s.router.Group("/api")
	protected.Use(auth.JWTMiddleware())
	{
		protected.POST("/container/create", handler.CreateContainer)
		protected.POST("/container/start", handler.CreateContainer)
		protected.POST("/container/exec", handler.Exec)
		// 添加其他需要验证的路由...
	}
}

func (s *Server) Run() error {
	return s.router.Run(s.config.ServerPort)
}
