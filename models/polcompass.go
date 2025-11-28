package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PolCompassController struct {
	DB *gorm.DB
}

type PolCompassReq struct {
	Field1Name  string     `json:"field1_name"`
	Field2Name  string     `json:"field2_name"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Questions   []Question `json:"questions"`
}

type Polcompass struct {
	gorm.Model
	Field1Name        string
	Field2Name        string
	Field1QuestionQty int
	Field2QuestionQty int
	Name              string
	Description       string
	Questions         []Question `json:"questions"`
}
type Question struct {
	ID           uint   `gorm:"primaryKey"` // Or gorm.Model is embedded
	Question     string `json:"question" gorm:"column:question;uniqueIndex:idx_polcompass_question"`
	Affects      string `json:"affects"`
	Direction    int    `json:"direction"`
	PolcompassID uint   `json:"-" gorm:"uniqueIndex:idx_polcompass_question"`
}

func (p *PolCompassController) GET(c *gin.Context) {
	reqID, isPresent := c.GetQuery("id")

	if !isPresent {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "You need to specify an id",
		})
		return
	}

	var polcompass Polcompass

	polcompassId64, err := strconv.ParseUint(reqID, 10, 64)
	polcompassId := uint(polcompassId64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "You need to specify an id",
		})
		return
	}

	p.DB.Preload("Questions").First(&polcompass, polcompassId)

	c.JSON(http.StatusOK, polcompass)

}

func (p *PolCompassController) First(c *gin.Context) {
	var polcompass Polcompass
	if err := p.DB.Preload("Questions").First(&polcompass).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "PolCompass not found",
		})
		return
	}
	c.JSON(http.StatusOK, polcompass)
}

func (p *PolCompassController) POST(c *gin.Context) {

	var req PolCompassReq
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad request for polcompass request field1_name string,field2_name string, questions [] ",
		})
		return
	}

	field1QuestionQty := 0
	field2QuestionQty := 0

	//Count questions for the frontend

	for _, q := range req.Questions {
		fmt.Print(q)
		if q.Affects == req.Field1Name {
			field1QuestionQty++
		} else if q.Affects == req.Field2Name {
			field2QuestionQty++
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "An unknown field was added in the questions : " + q.Affects + " fields names are : " + req.Field1Name + " and " + req.Field2Name,
			})
			return
		}
		q.ID = 0
	}
	newPolCompass := Polcompass{Field1Name: req.Field1Name, Field2Name: req.Field2Name, Field1QuestionQty: field1QuestionQty, Field2QuestionQty: field2QuestionQty, Name: req.Name, Description: req.Description}

	p.DB.Create(&newPolCompass)

	var questionsToSave []Question

	for _, q := range req.Questions {
		q.PolcompassID = newPolCompass.ID
		questionsToSave = append(questionsToSave, q)
	}

	result := p.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "question"}, {Name: "polcompass_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"direction": clause.Expr{SQL: "excluded.direction"}, "affects": clause.Expr{SQL: "excluded.affects"},
			"polcompass_id": clause.Expr{SQL: "excluded.polcompass_id"}}),
	}).Create(&questionsToSave)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error while saving questions to the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Added sucessfully to the database",
	})

}
