package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PolCompassIntegrationSuite struct {
	suite.Suite
	DB         *gorm.DB
	controller *PolCompassController
	router     *gin.Engine
	server     *httptest.Server
}

func (suite *PolCompassIntegrationSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	var err error
	suite.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	err = suite.DB.AutoMigrate(&Polcompass{}, &Question{})
	suite.Require().NoError(err)

	suite.controller = &PolCompassController{DB: suite.DB}
	suite.router = gin.New()

	suite.router.GET("/polcompass", suite.controller.GET)
	suite.router.GET("/polcompass/first", suite.controller.First)
	suite.router.POST("/polcompass", suite.controller.POST)

	suite.server = httptest.NewServer(suite.router)
}

func (suite *PolCompassIntegrationSuite) TearDownSuite() {
	suite.server.Close()
	sqlDB, _ := suite.DB.DB()
	sqlDB.Close()
}

func (suite *PolCompassIntegrationSuite) SetupTest() {
	// Clean database before each test
	suite.DB.Exec("DELETE FROM questions")
	suite.DB.Exec("DELETE FROM polcompasses")
}

func (suite *PolCompassIntegrationSuite) TestCompleteWorkflow() {
	// Step 1: Create a new polcompass
	createRequest := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Integration Test Compass",
		Description: "Testing complete workflow",
		Questions: []Question{
			{Question: "Should government regulate markets?", Affects: "Economic", Direction: -1},
			{Question: "Is free market best?", Affects: "Economic", Direction: 1},
			{Question: "Should abortion be legal?", Affects: "Social", Direction: 1},
			{Question: "Traditional values matter?", Affects: "Social", Direction: -1},
		},
	}

	jsonData, _ := json.Marshal(createRequest)
	resp, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var createResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResponse)
	assert.Equal(suite.T(), "Added sucessfully to the database", createResponse["message"])

	// Step 2: Retrieve the first polcompass
	resp, err = http.Get(suite.server.URL + "/polcompass/first")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var firstResponse Polcompass
	json.NewDecoder(resp.Body).Decode(&firstResponse)
	assert.Equal(suite.T(), "Integration Test Compass", firstResponse.Name)
	assert.Equal(suite.T(), "Economic", firstResponse.Field1Name)
	assert.Equal(suite.T(), "Social", firstResponse.Field2Name)
	assert.Equal(suite.T(), 2, firstResponse.Field1QuestionQty)
	assert.Equal(suite.T(), 2, firstResponse.Field2QuestionQty)
	assert.Len(suite.T(), firstResponse.Questions, 4)

	// Step 3: Retrieve specific polcompass by ID
	polcompassID := firstResponse.ID
	resp, err = http.Get(fmt.Sprintf("%s/polcompass?id=%d", suite.server.URL, polcompassID))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var getResponse Polcompass
	json.NewDecoder(resp.Body).Decode(&getResponse)
	assert.Equal(suite.T(), firstResponse.Name, getResponse.Name)
	assert.Equal(suite.T(), firstResponse.ID, getResponse.ID)
	assert.Len(suite.T(), getResponse.Questions, 4)

	// Verify questions are correct
	economicQuestions := 0
	socialQuestions := 0
	for _, question := range getResponse.Questions {
		if question.Affects == "Economic" {
			economicQuestions++
		} else if question.Affects == "Social" {
			socialQuestions++
		}
	}
	assert.Equal(suite.T(), 2, economicQuestions)
	assert.Equal(suite.T(), 2, socialQuestions)
}

func (suite *PolCompassIntegrationSuite) TestMultiplePolCompassCreation() {
	// Create first polcompass
	createRequest1 := PolCompassReq{
		Field1Name:  "Liberal",
		Field2Name:  "Conservative",
		Name:        "Political Compass 1",
		Description: "First compass",
		Questions: []Question{
			{Question: "Government size?", Affects: "Liberal", Direction: -1},
			{Question: "Individual freedom?", Affects: "Conservative", Direction: 1},
		},
	}

	jsonData1, _ := json.Marshal(createRequest1)
	resp1, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData1))
	suite.Require().NoError(err)
	defer resp1.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp1.StatusCode)

	// Create second polcompass
	createRequest2 := PolCompassReq{
		Field1Name:  "Authoritarian",
		Field2Name:  "Libertarian",
		Name:        "Political Compass 2",
		Description: "Second compass",
		Questions: []Question{
			{Question: "State control?", Affects: "Authoritarian", Direction: 1},
			{Question: "Personal liberty?", Affects: "Libertarian", Direction: -1},
			{Question: "Rule of law?", Affects: "Authoritarian", Direction: 1},
		},
	}

	jsonData2, _ := json.Marshal(createRequest2)
	resp2, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData2))
	suite.Require().NoError(err)
	defer resp2.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp2.StatusCode)

	// Verify first polcompass still accessible
	resp, err := http.Get(suite.server.URL + "/polcompass?id=1")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var firstCompass Polcompass
	json.NewDecoder(resp.Body).Decode(&firstCompass)
	assert.Equal(suite.T(), "Political Compass 1", firstCompass.Name)
	assert.Equal(suite.T(), "Liberal", firstCompass.Field1Name)
	assert.Len(suite.T(), firstCompass.Questions, 2)

	// Verify second polcompass accessible
	resp, err = http.Get(suite.server.URL + "/polcompass?id=2")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var secondCompass Polcompass
	json.NewDecoder(resp.Body).Decode(&secondCompass)
	assert.Equal(suite.T(), "Political Compass 2", secondCompass.Name)
	assert.Equal(suite.T(), "Authoritarian", secondCompass.Field1Name)
	assert.Len(suite.T(), secondCompass.Questions, 3)
}

func (suite *PolCompassIntegrationSuite) TestErrorHandlingWorkflow() {
	// Test creating polcompass with invalid data
	invalidRequest := map[string]interface{}{
		"field1_name": "Valid",
		"field2_name": "Valid",
		"questions": []map[string]interface{}{
			{
				"question":  "Test question",
				"affects":   "InvalidField",
				"direction": 1,
			},
		},
	}

	jsonData, _ := json.Marshal(invalidRequest)
	resp, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test getting non-existent polcompass
	resp, err = http.Get(suite.server.URL + "/polcompass?id=999")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	var response Polcompass
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(suite.T(), uint(0), response.ID)

	// Test getting polcompass without ID
	resp, err = http.Get(suite.server.URL + "/polcompass")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test getting first when no data exists
	resp, err = http.Get(suite.server.URL + "/polcompass/first")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

func (suite *PolCompassIntegrationSuite) TestQuestionConflictResolution() {
	// Create first polcompass with a question
	createRequest1 := PolCompassReq{
		Field1Name:  "Left",
		Field2Name:  "Right",
		Name:        "First Compass",
		Description: "Testing conflicts",
		Questions: []Question{
			{Question: "Shared Question", Affects: "Left", Direction: 1},
		},
	}

	jsonData1, _ := json.Marshal(createRequest1)
	resp1, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData1))
	suite.Require().NoError(err)
	defer resp1.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp1.StatusCode)

	// Create second polcompass with same question but different direction
	createRequest2 := PolCompassReq{
		Field1Name:  "Left",
		Field2Name:  "Right",
		Name:        "Second Compass",
		Description: "Testing conflicts resolution",
		Questions: []Question{
			{Question: "Shared Question", Affects: "Right", Direction: -1},
		},
	}

	jsonData2, _ := json.Marshal(createRequest2)
	resp2, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData2))
	suite.Require().NoError(err)
	defer resp2.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp2.StatusCode)

	// Verify both polcompasses exist and have correct questions
	resp, err := http.Get(suite.server.URL + "/polcompass?id=1")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var firstCompass Polcompass
	json.NewDecoder(resp.Body).Decode(&firstCompass)
	assert.Len(suite.T(), firstCompass.Questions, 1)
	assert.Equal(suite.T(), "Left", firstCompass.Questions[0].Affects)
	assert.Equal(suite.T(), 1, firstCompass.Questions[0].Direction)

	resp, err = http.Get(suite.server.URL + "/polcompass?id=2")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var secondCompass Polcompass
	json.NewDecoder(resp.Body).Decode(&secondCompass)
	assert.Len(suite.T(), secondCompass.Questions, 1)
	assert.Equal(suite.T(), "Right", secondCompass.Questions[0].Affects)
	assert.Equal(suite.T(), -1, secondCompass.Questions[0].Direction)
}

func (suite *PolCompassIntegrationSuite) TestLargeDatasetIntegration() {
	// Create a polcompass with many questions
	var questions []Question
	for i := 0; i < 100; i++ {
		field := "Economic"
		if i%2 == 0 {
			field = "Social"
		}
		questions = append(questions, Question{
			Question:  fmt.Sprintf("Large Dataset Question %d", i),
			Affects:   field,
			Direction: (i % 11) - 5, // Range from -5 to 5
		})
	}

	createRequest := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Large Dataset Compass",
		Description: "Testing with 100 questions",
		Questions:   questions,
	}

	jsonData, _ := json.Marshal(createRequest)
	resp, err := http.Post(suite.server.URL+"/polcompass", "application/json", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Retrieve and verify
	resp, err = http.Get(suite.server.URL + "/polcompass/first")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var response Polcompass
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(suite.T(), "Large Dataset Compass", response.Name)
	assert.Equal(suite.T(), 50, response.Field1QuestionQty) // Social (even indices)
	assert.Equal(suite.T(), 50, response.Field2QuestionQty) // Economic (odd indices)
	assert.Len(suite.T(), response.Questions, 100)
}

func TestPolCompassIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PolCompassIntegrationSuite))
}
