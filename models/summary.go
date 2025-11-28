package models

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SummaryResponse struct {
	NumberOfPages int `json:"numberOfPages"`
	Summaries     []Summary
}

type Summary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          uint   `json:"id"`
}

func (p *PolCompassController) Summary(c *gin.Context) {
	var summaries []Summary

	perPage, isPresent := c.GetQuery("perPage")
	if !isPresent {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please specify how many items you want to see per page",
		},
		)
		return
	}

	page, isPresent := c.GetQuery("page")

	if !isPresent {
		page = "1"
	}

	perPageInt, err := strconv.Atoi(perPage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "perPage must be a number",
		})
		return
	}

	if perPageInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "perPage must be greater than 0",
		})
		return
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "page must be a number",
		})
		return
	}

	if pageInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "page must be greater than 0",
		})
		return
	}

	p.DB.Model(&Polcompass{}).Limit(perPageInt).Offset((pageInt - 1) * perPageInt).Select("id, name, description").Where("name != '' and description != ''").Find(&summaries)

	var num_summaries int64
	p.DB.Model(&Polcompass{}).Select("id, name, description").Where("name != '' and description != ''").Count(&num_summaries)
	numberOfPages := int(math.Ceil(float64(num_summaries) / float64(perPageInt)))

	summariesResponse := SummaryResponse{
		NumberOfPages: numberOfPages,
		Summaries:     summaries,
	}

	c.JSON(http.StatusOK, summariesResponse)

}
