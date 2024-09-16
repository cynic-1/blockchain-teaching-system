package server

import (
	"github.com/cynic-1/blockchain-teaching-system/internal/api"
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

	s.router.POST("/register", handler.Register)
	s.router.POST("/login", handler.Login)
	s.router.POST("/container", handler.CreateContainer)
	// 添加其他路由...
}

func (s *Server) Run() error {
	return s.router.Run(s.config.ServerPort)
}
