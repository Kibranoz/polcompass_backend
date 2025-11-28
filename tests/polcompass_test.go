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

type PolCompassTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	controller *PolCompassController
	router     *gin.Engine
}

func (suite *PolCompassTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *PolCompassTestSuite) SetupTest() {
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
}

func (suite *PolCompassTestSuite) TearDownTest() {
	sqlDB, _ := suite.DB.DB()
	sqlDB.Close()
}

// Test GET method
func (suite *PolCompassTestSuite) TestGET_MissingID() {
	req, _ := http.NewRequest("GET", "/polcompass", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "You need to specify an id", response["message"])
}

func (suite *PolCompassTestSuite) TestGET_InvalidID() {
	req, _ := http.NewRequest("GET", "/polcompass?id=invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "You need to specify an id", response["message"])
}

func (suite *PolCompassTestSuite) TestGET_ValidID() {
	// Create test data
	polcompass := Polcompass{
		Field1Name:        "Economic",
		Field2Name:        "Social",
		Field1QuestionQty: 1,
		Field2QuestionQty: 1,
		Name:              "Test Compass",
		Description:       "Test Description",
	}
	suite.DB.Create(&polcompass)

	questions := []Question{
		{Question: "Test Question 1", Affects: "Economic", Direction: 1, PolcompassID: polcompass.ID},
		{Question: "Test Question 2", Affects: "Social", Direction: -1, PolcompassID: polcompass.ID},
	}
	suite.DB.Create(&questions)

	req, _ := http.NewRequest("GET", "/polcompass?id=1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response Polcompass
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Test Compass", response.Name)
	assert.Equal(suite.T(), "Economic", response.Field1Name)
	assert.Equal(suite.T(), "Social", response.Field2Name)
	assert.Len(suite.T(), response.Questions, 2)
}

// Test First method
func (suite *PolCompassTestSuite) TestFirst_NoData() {
	req, _ := http.NewRequest("GET", "/polcompass/first", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "PolCompass not found", response["message"])
}

func (suite *PolCompassTestSuite) TestFirst_WithData() {
	// Create test data
	polcompass := Polcompass{
		Field1Name:        "Economic",
		Field2Name:        "Social",
		Field1QuestionQty: 1,
		Field2QuestionQty: 1,
		Name:              "First Compass",
		Description:       "First Description",
	}
	suite.DB.Create(&polcompass)

	questions := []Question{
		{Question: "First Question", Affects: "Economic", Direction: 1, PolcompassID: polcompass.ID},
	}
	suite.DB.Create(&questions)

	req, _ := http.NewRequest("GET", "/polcompass/first", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response Polcompass
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "First Compass", response.Name)
	assert.Equal(suite.T(), "Economic", response.Field1Name)
	assert.Len(suite.T(), response.Questions, 1)
}

// Test POST method
func (suite *PolCompassTestSuite) TestPOST_InvalidJSON() {
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response["message"], "Bad request for polcompass request")
}

func (suite *PolCompassTestSuite) TestPOST_UnknownField() {
	requestBody := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Test Compass",
		Description: "Test Description",
		Questions: []Question{
			{Question: "Test Question", Affects: "Unknown", Direction: 1},
		},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response["message"], "An unknown field was added in the questions : Unknown")
}

func (suite *PolCompassTestSuite) TestPOST_MissingFieldNames() {
	requestBody := PolCompassReq{
		Field1Name:  "",
		Field2Name:  "Social",
		Name:        "Test Compass",
		Description: "Test Description",
		Questions: []Question{
			{Question: "Test Question", Affects: "Social", Direction: 1},
		},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response["message"], "An unknown field was added in the questions")
}

func (suite *PolCompassTestSuite) TestPOST_ValidRequest() {
	requestBody := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Test Compass",
		Description: "Test Description",
		Questions: []Question{
			{Question: "Economic Question 1", Affects: "Economic", Direction: 1},
			{Question: "Economic Question 2", Affects: "Economic", Direction: -1},
			{Question: "Social Question 1", Affects: "Social", Direction: 1},
		},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Added sucessfully to the database", response["message"])

	// Verify data was saved correctly
	var savedPolcompass Polcompass
	suite.DB.Preload("Questions").First(&savedPolcompass)
	assert.Equal(suite.T(), "Test Compass", savedPolcompass.Name)
	assert.Equal(suite.T(), "Economic", savedPolcompass.Field1Name)
	assert.Equal(suite.T(), "Social", savedPolcompass.Field2Name)
	assert.Equal(suite.T(), 2, savedPolcompass.Field1QuestionQty)
	assert.Equal(suite.T(), 1, savedPolcompass.Field2QuestionQty)
	assert.Len(suite.T(), savedPolcompass.Questions, 3)

	// Verify question directions are saved correctly
	for _, question := range savedPolcompass.Questions {
		if question.Question == "Economic Question 1" {
			assert.Equal(suite.T(), 1, question.Direction)
		} else if question.Question == "Economic Question 2" {
			assert.Equal(suite.T(), -1, question.Direction)
		} else if question.Question == "Social Question 1" {
			assert.Equal(suite.T(), 1, question.Direction)
		}
	}
}

func (suite *PolCompassTestSuite) TestPOST_EmptyQuestions() {
	requestBody := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Empty Compass",
		Description: "No Questions",
		Questions:   []Question{},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Added sucessfully to the database", response["message"])

	// Verify data was saved correctly
	var savedPolcompass Polcompass
	suite.DB.Preload("Questions").First(&savedPolcompass)
	assert.Equal(suite.T(), "Empty Compass", savedPolcompass.Name)
	assert.Equal(suite.T(), 0, savedPolcompass.Field1QuestionQty)
	assert.Equal(suite.T(), 0, savedPolcompass.Field2QuestionQty)
	assert.Len(suite.T(), savedPolcompass.Questions, 0)
}

func (suite *PolCompassTestSuite) TestPOST_DuplicateQuestions() {
	// First create a polcompass with questions
	requestBody1 := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "First Compass",
		Description: "First Description",
		Questions: []Question{
			{Question: "Duplicate Question", Affects: "Economic", Direction: 1},
		},
	}

	jsonData1, _ := json.Marshal(requestBody1)
	req1, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)

	assert.Equal(suite.T(), http.StatusOK, w1.Code)

	// Now create another polcompass with the same question but different direction
	requestBody2 := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Second Compass",
		Description: "Second Description",
		Questions: []Question{
			{Question: "Duplicate Question", Affects: "Economic", Direction: -1},
		},
	}

	jsonData2, _ := json.Marshal(requestBody2)
	req2, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	// Verify both polcompasses exist
	var polcompasses []Polcompass
	suite.DB.Find(&polcompasses)
	assert.Len(suite.T(), polcompasses, 2)
}

func (suite *PolCompassTestSuite) TestPOST_NegativeDirections() {
	requestBody := PolCompassReq{
		Field1Name:  "Left",
		Field2Name:  "Right",
		Name:        "Direction Test Compass",
		Description: "Testing negative and positive directions",
		Questions: []Question{
			{Question: "Negative Direction Question", Affects: "Left", Direction: -5},
			{Question: "Zero Direction Question", Affects: "Right", Direction: 0},
			{Question: "Positive Direction Question", Affects: "Left", Direction: 3},
		},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify directions are preserved
	var savedPolcompass Polcompass
	suite.DB.Preload("Questions").First(&savedPolcompass)

	for _, question := range savedPolcompass.Questions {
		switch question.Question {
		case "Negative Direction Question":
			assert.Equal(suite.T(), -5, question.Direction)
		case "Zero Direction Question":
			assert.Equal(suite.T(), 0, question.Direction)
		case "Positive Direction Question":
			assert.Equal(suite.T(), 3, question.Direction)
		}
	}
}

func (suite *PolCompassTestSuite) TestGET_NonExistentID() {
	req, _ := http.NewRequest("GET", "/polcompass?id=999", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response Polcompass
	json.Unmarshal(w.Body.Bytes(), &response)
	// Should return empty polcompass when ID doesn't exist
	assert.Equal(suite.T(), uint(0), response.ID)
}

func (suite *PolCompassTestSuite) TestPOST_LargeDataSet() {
	var questions []Question
	for i := 0; i < 50; i++ {
		field := "Economic"
		if i%2 == 0 {
			field = "Social"
		}
		questions = append(questions, Question{
			Question:  fmt.Sprintf("Question %d", i),
			Affects:   field,
			Direction: i - 25, // Range from -25 to 24
		})
	}

	requestBody := PolCompassReq{
		Field1Name:  "Economic",
		Field2Name:  "Social",
		Name:        "Large Dataset Compass",
		Description: "Testing with many questions",
		Questions:   questions,
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/polcompass", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify counts are correct
	var savedPolcompass Polcompass
	suite.DB.Preload("Questions").First(&savedPolcompass)
	assert.Equal(suite.T(), 25, savedPolcompass.Field1QuestionQty) // Social questions (even indices)
	assert.Equal(suite.T(), 25, savedPolcompass.Field2QuestionQty) // Economic questions (odd indices)
	assert.Len(suite.T(), savedPolcompass.Questions, 50)
}

func TestPolCompassSuite(t *testing.T) {
	suite.Run(t, new(PolCompassTestSuite))
}
