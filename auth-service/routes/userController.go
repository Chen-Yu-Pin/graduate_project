package routes

import (
	"auth-service/models"
	"auth-service/util"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
	Mongo *models.Mongo
	Redis *models.Redis
}

func NewUserController() (*UserController, error) {
	mongo, err := models.NewMongo()
	if err != nil {
		return nil, err
	}
	redis := models.NewRedisConnect()
	return &UserController{
		Mongo: mongo,
		Redis: redis,
	}, nil
}

func (n *UserController) VerifyMail(c *gin.Context) {
	var data map[string]interface{}
	if err := c.Bind(&data); err != nil {
		log.Println("what's going on")
		c.JSON(http.StatusBadRequest, gin.H{"result": err})
		return
	}
	mail := fmt.Sprintf("%v", data["user-mail"])

	verifyCode := generateVerifyCode()
	err := n.Redis.SetVerifyCode(mail, verifyCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err})
		return
	}
	err = util.SendVerifyCode(verifyCode, mail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"status": "success",
	})
}
func (n *UserController) CreateUser(c *gin.Context) {
	type data struct {
		UserName           string `json:"user-name"`
		UserPassword       string `json:"user-password"`
		UserMail           string `json:"user-email"`
		VerifyCode         string `json:"email-verification-code"`
		UserHeadshotNumber string `json:"user-headshot-number"`
	}
	var requestData data
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "bindingError"})
		return
	}
	log.Println(requestData)
	verifyCode, err := n.Redis.GetVerifyCode(requestData.UserMail)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"result": "verifyError"})
		return
	}
	if requestData.VerifyCode == *verifyCode {
		insertData := &models.UserInfo{
			UserEstablishedDate: time.Now(),
			UserID:              bson.NewObjectId(),
			UserName:            requestData.UserName,
			UserPassword:        requestData.UserPassword,
			UserHeadshotNumber:  requestData.UserHeadshotNumber,
			UserMail:            requestData.UserMail,
			PersonalOperateHistory: models.PersonalOperateHistorys{
				NumberOfReceivedLikes:        0,
				NumberOfReleasedComments:     0,
				NumberOfCollectBoards:        0,
				NumberOfLikeBoards:           0,
				NumberOfLikeComments:         0,
				NumberOfLaunchPetitionBoards: 0,
				NumberOfSupportSigningBoards: 0,
			},
			CollectedBoard:        []models.CollectedBoards{},
			LikededBoard:          []models.LikedBoards{},
			SupportedSigningBoard: []models.SupportedSigningBoards{},
		}
		if err := n.Mongo.InsertUser(insertData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "insertError"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "success",
			"user-id": insertData.UserID,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"result": "verify error"})
		return
	}
}
func (n *UserController) Login(c *gin.Context) {
	type User struct {
		UserAccount  string `json:"user-account"`
		UserPassword string `json:"user-password"`
	}
	var data User
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bindingError"})
		return
	}
	log.Println(data)
	var account *models.UserInfo
	success, account, err := n.Mongo.LoginCheck(data.UserAccount, data.UserPassword)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "verify error"})
		return
	}
	if !success {
		log.Println("not success")
		c.JSON(http.StatusBadRequest, gin.H{"status": "verify error"})
		return
	} else {
		token, err := util.GenToken(data.UserAccount)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "token generated error"})
			return
		}
		if err := n.Redis.JoinToTokenList(account.UserID.Hex(), token); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "insert error"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"status":               "success",
			"user-name":            account.UserName,
			"user-headshot-number": account.UserHeadshotNumber,
			"user-email":           account.UserMail,
			"user-id":              account.UserID,
			"user-token":           token,
		})
	}
}
func (n *UserController) Logout(c *gin.Context) {
	userID := c.Request.Header.Get("user-id")
	authHeader := c.Request.Header.Get("Authorization")
	err := n.Redis.DeleteToken(userID, authHeader)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "logout error"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"status": "success",
	})
}
func (n *UserController) SearchUser(c *gin.Context) {

	data := c.Request.Header.Get("user-id")

	person, err := n.Mongo.SearchUser(data)
	if err != nil {
		log.Println("search error", err)
		c.JSON(http.StatusBadRequest, gin.H{"result": err})
		return
	}
	log.Println(*person)
	jsondata, err := json.Marshal(&person)
	log.Println(jsondata)
	if err != nil {
		log.Println("switch error")
		c.JSON(http.StatusBadRequest, gin.H{"result": err})
		return
	}
	c.JSON(http.StatusAccepted, *person)
}

func (n *UserController) JWTAuthMiddleware(c *gin.Context) {

	// var data map[string]interface{}
	// if err := c.Bind(&data); err != nil {
	// 	log.Println("what's going on")
	// 	log.Println(err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"result": err})
	// 	return
	// }
	userID := c.Request.Header.Get("user-id")
	// Get token from Header.Authorization field.
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": -1,
			"msg":  "Authorization is null",
		})
		c.Abort()
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": -1,
			"msg":  "Format of Authorization is wrong",
		})
		c.Abort()
		return
	}
	// parts[0] is Bearer, parts is token.
	err := n.Redis.SearchToken(userID, parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": -1,
			"msg":  "nil token",
		})
		c.Abort()
		return
	}
	err = util.ParseToken(parts[1])
	if err != nil {
		log.Println(err)
		if err.Error() == "token expired" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "token expired",
			})
			c.Abort()
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": -1,
			"msg":  "Invalid Token.",
		})
		c.Abort()
		return
	}
	// Store Account info into Context
	c.Set("req-data", userID)
	// After that, we can get Account info from c.Get("account")
	c.Next()

}

func generateVerifyCode() string {
	rand.Seed(time.Now().UnixNano())
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 6)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
