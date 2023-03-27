package main

import (
	"auth-service/auth"
	board "auth-service/boardgRPCoperate"
	petition "auth-service/petitiongRPC"
	"auth-service/routes"
	"auth-service/util"
	"log"
	"net"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
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
	UserController, err := routes.NewUserController()
	if err != nil {
		log.Println(err)
		return
	}
	go createGrpcServer(*UserController)

	server.POST("/post-require-verification-code", UserController.VerifyMail)
	server.POST("/post-register", UserController.CreateUser)
	server.POST("/post-login", UserController.Login)
	server.POST("/post-query-personal-infos", UserController.JWTAuthMiddleware, UserController.SearchUser)
	server.POST("/post-logout", UserController.JWTAuthMiddleware, UserController.Logout)
	server.Run(":80")
}

func createGrpcServer(UserController routes.UserController) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, &util.AuthServer{RedisModel: *UserController.Redis})
	board.RegisterBoardServiceServer(s, &util.BoardServer{MongoModel: *UserController.Mongo})
	petition.RegisterPetitionServiceServer(s, &util.PetitionServer{MongoModel: *UserController.Mongo})
	log.Printf("gRPC Server started on port %s", "50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
