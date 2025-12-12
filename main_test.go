package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/database"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/handlers"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	// Configuração do banco de dados de teste
	db, err := database.ConnectTest()
	if err != nil {
		t.Fatal("Falha ao conectar ao banco de dados de teste:", err)
	}
	defer database.CloseTest(db)

	// Dados do usuário para o teste
	userData := map[string]string{
		"email":    "testuser@example.com",
		"password": "testpassword",
	}
	jsonBody, _ := json.Marshal(userData)

	// Cria uma requisição de teste
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	// Cria um ResponseRecorder para gravar a resposta
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.RegisterHandler(db))

	// Executa a requisição
	handler.ServeHTTP(rr, req)

	// Verifica o código de status
	assert.Equal(t, http.StatusCreated, rr.Code, "O código de status esperado era 201")

	// Verifica o corpo da resposta
	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	assert.Equal(t, "User created successfully", responseBody["message"])
	assert.NotNil(t, responseBody["user"])
	assert.NotNil(t, responseBody["forge_api_key"])
}

func TestLoginHandler(t *testing.T) {
	// Configuração do banco de dados de teste
	db, err := database.ConnectTest()
	if err != nil {
		t.Fatal("Falha ao conectar ao banco de dados de teste:", err)
	}
	defer database.CloseTest(db)

	// Cria um usuário para o teste usando o RegisterHandler
	userData := map[string]string{
		"email":    "testuser@example.com",
		"password": "testpassword",
	}
	jsonBody, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	registerHandler := http.HandlerFunc(handlers.RegisterHandler(db))
	registerHandler.ServeHTTP(rr, req)

	// Dados do usuário para o teste
	loginData := map[string]string{
		"email":    "testuser@example.com",
		"password": "testpassword",
	}
	jsonLoginBody, _ := json.Marshal(loginData)

	// Cria uma requisição de teste
	loginReq, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonLoginBody))
	if err != nil {
		t.Fatal(err)
	}

	// Cria um ResponseRecorder para gravar a resposta
	loginRR := httptest.NewRecorder()
	loginHandler := http.HandlerFunc(handlers.LoginHandler(db))

	// Executa a requisição
	loginHandler.ServeHTTP(loginRR, loginReq)

	// Verifica o código de status
	assert.Equal(t, http.StatusOK, loginRR.Code, "O código de status esperado era 200")

	// Verifica o corpo da resposta
	var responseBody map[string]interface{}
	err = json.Unmarshal(loginRR.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	assert.Equal(t, "Logged in successfully", responseBody["message"])
	assert.NotNil(t, responseBody["token"])
}