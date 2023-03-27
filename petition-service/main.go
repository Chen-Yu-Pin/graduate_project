package main

import (
	"log"
	"petition-service/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "user-id"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	petitionController, err := routes.NewController()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	server.GET("/get-signing-achieved-board", petitionController.GetBoard)
	server.POST("/post-support-signing-board", petitionController.SignBoard)
	server.POST("/post-initiate-board", petitionController.InitBoard)
	server.GET("/get-signing-board-infos/:board-id", petitionController.GetBoardsInfo)
	server.Run(":80")
}
