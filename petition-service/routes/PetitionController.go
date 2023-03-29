package routes

import (
	"context"
	"log"
	"net/http"
	"petition-service/auth"
	"petition-service/models"
	petition "petition-service/petitiongRPC"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PetitionController struct {
	MongoModel *models.Mongo
}

func NewController() (*PetitionController, error) {
	mongo, err := models.NewMongo()
	if err != nil {
		return nil, err
	}
	controller := &PetitionController{
		MongoModel: mongo,
	}
	return controller, nil
}
func (p *PetitionController) GetBoard(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	if AuthMethod(authHeader, user_id) {
		result, err := p.MongoModel.GetAllBoard(user_id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "getError",
			})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"signing-boards":           result.SigningBoard,
			"recently-achieved-boards": result.AchievedBoard,
		})
		return
	} else {
		result, err := p.MongoModel.GetAllBoard("")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "getError",
			})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"signing-boards":           result.SigningBoard,
			"recently-achieved-boards": result.AchievedBoard,
		})
	}
}
func (p *PetitionController) SignBoard(c *gin.Context) {
	type requestBody struct {
		BoardID   string `json:"signing-board-id"`
		IsSupport bool   `json:"is-supported"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "BindingError",
		})
		return
	}
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")

	if AuthMethod(authHeader, user_id) {
		err := p.MongoModel.SupportBoard(user_id, requestData.BoardID, requestData.IsSupport)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "SigningError",
			})
			return
		}
		conn, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "createRequestError",
			})
			return
		}
		defer conn.Close()
		connection := petition.NewPetitionServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := connection.AddSigningBoard(ctx, &petition.SignBoard{
			SignBoardID: requestData.BoardID,
			UserID:      user_id,
			IsSign:      requestData.IsSupport,
		})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
		if res.Result {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "success",
			})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "authError",
		})
	}
}
func (p *PetitionController) InitBoard(c *gin.Context) {
	type requestBody struct {
		BoardTitle      string `json:"board-title"`
		BoardMotivation string `json:"board-motivation"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "BindingError",
		})
		return
	}
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	if AuthMethod(authHeader, user_id) {
		err := p.MongoModel.InitBoard(requestData.BoardTitle, requestData.BoardMotivation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "InitError",
			})
			return
		}
		conn, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "createRequestError",
			})
			return
		}
		defer conn.Close()
		connection := petition.NewPetitionServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := connection.LaunchPetitionBoard(ctx, &petition.LaunchBoard{
			UserID: user_id,
		})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
		if res.Result {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "success",
			})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "authError",
		})
	}
}
func (p *PetitionController) GetBoardsInfo(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	boardID := c.Param("board-id")
	if AuthMethod(authHeader, user_id) {
		result, err := p.MongoModel.GetBoard(boardID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "search error",
			})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"board-title":       result.BoardTitle,
			"number-of-signers": result.NumberOfSigners,
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "authError",
		})
	}
}
func AuthMethod(authHeader, user_id string) bool {
	if authHeader == "" {
		return false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return false
	}
	conn, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Println(err)
		return false
	}
	defer conn.Close()
	c := auth.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.AuthAccount(ctx, &auth.AuthRequest{
		AuthEntry: &auth.Auth{
			Token:  parts[1],
			UserID: user_id,
		},
	})
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println(res.Result)
	return res.Result
}
