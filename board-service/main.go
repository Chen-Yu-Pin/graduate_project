package main

import (
	"board-service/routes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	server.StaticFS("/image", http.Dir("./img"))
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "user-id", "user-name"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	boardController, err := routes.NewBoardController()
	if err != nil {
		log.Println(err)
		return
	}
	popularController, err := routes.NewPopularController()
	if err != nil {
		log.Println(err)
		return
	}
	server.GET("/get-popular-keywords", popularController.PopulerTags)
	server.GET("/get-popular-boards", popularController.PopularBoards)
	server.GET("/get-read-board/:boardID", boardController.GetBoard)
	server.POST("/post-comment", boardController.PostComment)
	server.POST("/post-collect-board", boardController.CollectBoard)
	server.POST("/post-like-board", boardController.LikeBoard)
	server.POST("/post-like-comment", boardController.LikeComment)
	server.POST("/post-report-comment", boardController.ReportComment)
	server.POST("/post-search-board", boardController.SearchBoard)
	server.GET("/get-photo/:photo", func(c *gin.Context) {
		photo := c.Param("photo")
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": err,
			})
			return
		}
		log.Println(path)
		file, err := os.Open(path + "app/img/" + photo)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": err,
			})
			return
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": err,
			})
			return
		}
		c.Data(200, "image/png", b)
	})
	server.Run(":80")
}
