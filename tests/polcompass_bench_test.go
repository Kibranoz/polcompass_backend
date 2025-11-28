package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func BenchmarkPolCompassPOST(b *testing.B) {
	gin.SetMode(gin.TestMode)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&Polcompass{}, &Question{})

	controller := &PolCompassController{DB: db}
	router := gin.New()
	router.POST("/polcompass", controller.POST)

	requestBody := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Benchmark Compass",
		Description: "Performance testing",
		Questions: []Question{
			{Question: "Economic Question 1", Affects: "Economic", Direction: 1},
			{Question: "Economic Question 2", Affects: "Economic", Direction: -1},
			{Question: "Social Question 1", Affects: "Social", Direction: 1},
			{Question: "Social Question 2", Affects: "Social", Direction: -1},
		},
	}

	jsonData, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPolCompassGET(b *testing.B) {
	gin.SetMode(gin.TestMode)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&Polcompass{}, &Question{})

	// Setup test data
	polcompass := Polcompass{
		Field1Name:        "Economic",
		Field2Name:        "Social",
		Field1QuestionQty: 2,
		Field2QuestionQty: 2,
		Name:              "Benchmark Compass",
		Description:       "Performance testing",
	}
	db.Create(&polcompass)

	questions := []Question{
		{Question: "Test Question 1", Affects: "Economic", Direction: 1, PolcompassID: polcompass.ID},
		{Question: "Test Question 2", Affects: "Social", Direction: -1, PolcompassID: polcompass.ID},
	}
	db.Create(&questions)

	controller := &PolCompassController{DB: db}
	router := gin.New()
	router.GET("/polcompass", controller.GET)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/polcompass?id=1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPolCompassFirst(b *testing.B) {
	gin.SetMode(gin.TestMode)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&Polcompass{}, &Question{})

	// Setup test data
	polcompass := Polcompass{
		Field1Name:        "Economic",
		Field2Name:        "Social",
		Field1QuestionQty: 1,
		Field2QuestionQty: 1,
		Name:              "Benchmark Compass",
		Description:       "Performance testing",
	}
	db.Create(&polcompass)

	questions := []Question{
		{Question: "Test Question", Affects: "Economic", Direction: 1, PolcompassID: polcompass.ID},
	}
	db.Create(&questions)

	controller := &PolCompassController{DB: db}
	router := gin.New()
	router.GET("/polcompass/first", controller.First)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/polcompass/first", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPolCompassPOSTLargeDataset(b *testing.B) {
	gin.SetMode(gin.TestMode)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&Polcompass{}, &Question{})

	controller := &PolCompassController{DB: db}
	router := gin.New()
	router.POST("/polcompass", controller.POST)

	// Create large dataset
	var questions []Question
	for i := 0; i < 100; i++ {
		field := "Economic"
		if i%2 == 0 {
			field = "Social"
		}
		questions = append(questions, Question{
			Question:  "Benchmark Question " + string(rune(i)),
			Affects:   field,
			Direction: i - 50,
		})
	}

	requestBody := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Large Benchmark Compass",
		Description: "Performance testing with many questions",
		Questions:   questions,
	}

	jsonData, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
