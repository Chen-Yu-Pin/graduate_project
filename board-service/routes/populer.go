package routes

import (
	"board-service/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PopularController struct {
	MongoModel *models.DataModel
}

func NewPopularController() (*PopularController, error) {
	boardmodel, err := models.NewBoardModel()
	if err != nil {
		return nil, err
	}
	return &PopularController{
		MongoModel: boardmodel,
	}, nil
}

func (p *PopularController) PopulerTags(c *gin.Context) {
	tags, err := p.MongoModel.GetPopularTags()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"tags": tags,
	})
}
func (p *PopularController) PopularBoards(c *gin.Context) {
	boards, err := p.MongoModel.GetPopularBoards()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"data": boards,
	})
}
