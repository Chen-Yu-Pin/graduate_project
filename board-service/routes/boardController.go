package routes

import (
	"board-service/auth"
	board "board-service/boardgRPCoperate"
	"board-service/models"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BoardController struct {
	boardModel *models.DataModel
}

func NewBoardController() (*BoardController, error) {
	boardmodel, err := models.NewBoardModel()
	if err != nil {
		return nil, err
	}
	boardmodel.SetFakeValue()
	return &BoardController{
		boardModel: boardmodel,
	}, nil
}
func (b *BoardController) GetBoard(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	authUser := c.Request.Header.Get("user-id")
	boardID := c.Param("boardID")
	log.Println(boardID)
	id := uuid.FromStringOrNil(boardID)
	result, err := b.boardModel.GetParticularBoard(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
	}
	if AuthMethod(authHeader, authUser) {
		b.boardModel.CheckPersonalInfoForBoard(authUser, result)
		c.JSON(http.StatusAccepted, gin.H{
			"board": result,
		})
	} else {
		c.JSON(http.StatusAccepted, gin.H{
			"board": result,
		})
	}
}
func (b *BoardController) PostComment(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	user_name := c.Request.Header.Get("user-name")
	type Comment struct {
		CommentUserID string `json:"commented-user-id"`
		Content       string `json:"content"`
	}
	type RequestBody struct {
		PositionID  string  `json:"post-position-id"`
		CommentInfo Comment `json:"comment"`
	}
	var requestData RequestBody
	if AuthMethod(authHeader, user_id) {
		err := c.ShouldBindJSON(&requestData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "bindingError",
			})
			return
		}
		log.Println(requestData)
		storeData := &models.Comment{

			PositionID:      uuid.FromStringOrNil(requestData.PositionID),
			CommentedUserID: requestData.CommentInfo.CommentUserID,
			CommentUserName: user_name,
			CommentedDate:   time.Now(),
			ReportedCount:   0,
			Content:         requestData.CommentInfo.Content,
			NumOfLiker:      0,
		}
		err = b.boardModel.PostComment(storeData)
		log.Println(storeData)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "insertError",
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
		connection := board.NewBoardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := connection.ReleaseComment(ctx, &board.ReleaseCommentInfo{
			UserID:    user_id,
			CommentID: storeData.ID.String(),
		})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
		log.Println(res.Result)
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
		return
	}
}
func (b *BoardController) CollectBoard(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	type requestBody struct {
		BoardID     string `json:"board-id"`
		IsCollected bool   `json:"is-collected"`
		BoardTitle  string `json:"board-title"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "bindingError",
		})
		return
	}
	log.Println(requestData)
	id := uuid.FromStringOrNil(requestData.BoardID)
	if AuthMethod(authHeader, user_id) {
		b.boardModel.CollectBoard(&user_id, &id, requestData.IsCollected)
		conn, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "createRequestError",
			})
			return
		}
		defer conn.Close()
		connection := board.NewBoardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := connection.CollectBoard(ctx, &board.CollectBoardInfo{
			UserID:     user_id,
			IsCollect:  requestData.IsCollected,
			BoardID:    requestData.BoardID,
			BoardTitle: requestData.BoardTitle,
		})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
		log.Println(res.Result)
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
		return
	}
}
func (b *BoardController) LikeBoard(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	type requestBody struct {
		BoardID string `json:"board-id"`
		IsLike  bool   `json:"is-like"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "bindingError",
		})
		return
	}
	log.Println(requestData)
	id := uuid.FromStringOrNil(requestData.BoardID)
	if AuthMethod(authHeader, user_id) {
		b.boardModel.LikeBoard(&user_id, &id, requestData.IsLike)
		conn, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "createRequestError",
			})
			return
		}
		defer conn.Close()
		connection := board.NewBoardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := connection.LikeBoard(ctx, &board.LikeBoardInfo{
			UserID:  user_id,
			IsLiked: requestData.IsLike,
			BoardID: requestData.BoardID,
		})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
		log.Println(res.Result)
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
		return
	}
}

func (b *BoardController) LikeComment(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	type requestBody struct {
		LikePositionID string `json:"like-position-id"`
		LIkeCommentID  string `json:"like-comment-id"`
		IsLike         bool   `json:"is-like"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "bindingError",
		})
		return
	}
	log.Println(requestData)
	id := uuid.FromStringOrNil(requestData.LIkeCommentID)
	if AuthMethod(authHeader, user_id) {
		comment, err := b.boardModel.LikeComment(&user_id, &id, requestData.IsLike)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusAccepted, gin.H{
				"status": "updateError",
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
		connection := board.NewBoardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := connection.LikeComment(ctx, &board.LikeCommentInfo{
			GiverID:     user_id,
			IsGiverLike: requestData.IsLike,
			ReceiverID:  comment.CommentedUserID,
			CommentID:   comment.ID.String(),
		})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "responseError",
			})
			return
		}
		log.Println(res.Result)
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
		return
	}
}
func (b *BoardController) ReportComment(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	user_id := c.Request.Header.Get("user-id")
	type requestBody struct {
		ReportCommentID string `json:"report-comment-id"`
		IsReported      bool   `json:"is-reported"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "bindingError",
		})
		return
	}
	log.Println(requestData)
	id := uuid.FromStringOrNil(requestData.ReportCommentID)
	if AuthMethod(authHeader, user_id) {
		err := b.boardModel.ReportComment(&user_id, &id, requestData.IsReported)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "reportError",
			})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"status": "success",
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "authError",
		})
		return
	}
}

func (b *BoardController) SearchBoard(c *gin.Context) {
	type requestBody struct {
		Keyword string `json:"keyword"`
	}
	var requestData requestBody
	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "bindingError",
		})
		return
	}
	log.Println(requestData)
	KeywordList := strings.Split(requestData.Keyword, " ")
	result, err := b.boardModel.SearchBoardByKeywordsOrTags(KeywordList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "searchError",
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"data": result,
	})
}

func (b *BoardController) GetPhoto(c *gin.Context) {
	photo := c.Param("photo")
	log.Println(photo)
	c.File("/public/" + photo + ".png")
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
