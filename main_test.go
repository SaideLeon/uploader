package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/database"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/handlers"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	config.LoadConfig()
	db, err := database.ConnectTest()
	if err != nil {
		t.Fatal("Falha ao conectar ao banco de dados de teste:", err)
	}
	defer database.CloseTest(db)

	userData := map[string]string{
		"name":            "Test User",
		"email":           "testuser@example.com",
		"password":        "Password@123",
		"whatsapp_number": "+1234567890",
	}
	jsonBody, _ := json.Marshal(userData)

	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.RegisterHandler(db))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "O código de status esperado era 201")

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	assert.Equal(t, "User created successfully", responseBody["message"])
	assert.NotNil(t, responseBody["user"])
	assert.NotNil(t, responseBody["forge_api_key"])
}

func TestLoginHandler(t *testing.T) {
	config.LoadConfig()
	db, err := database.ConnectTest()
	if err != nil {
		t.Fatal("Falha ao conectar ao banco de dados de teste:", err)
	}
	defer database.CloseTest(db)

	// Registrar um usuário para o teste
	registerData := map[string]string{
		"name":            "Test User",
		"email":           "testuser@example.com",
		"password":        "Password@123",
		"whatsapp_number": "+1234567890",
	}
	jsonRegisterBody, _ := json.Marshal(registerData)
	registerReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonRegisterBody))
	registerRR := httptest.NewRecorder()
	handlers.RegisterHandler(db)(registerRR, registerReq)
	assert.Equal(t, http.StatusCreated, registerRR.Code)

	// Testar login com email
	loginDataEmail := map[string]string{
		"email":    "testuser@example.com",
		"password": "Password@123",
	}
	jsonLoginBodyEmail, _ := json.Marshal(loginDataEmail)
	loginReqEmail, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonLoginBodyEmail))
	loginRREmail := httptest.NewRecorder()
	handlers.LoginHandler(db)(loginRREmail, loginReqEmail)

	assert.Equal(t, http.StatusOK, loginRREmail.Code, "O código de status esperado para login com email era 200")
	var responseBodyEmail map[string]interface{}
	json.Unmarshal(loginRREmail.Body.Bytes(), &responseBodyEmail)
	assert.Equal(t, "Logged in successfully", responseBodyEmail["message"])
	assert.NotNil(t, responseBodyEmail["token"])
}
