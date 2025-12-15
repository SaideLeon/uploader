// @title MidiaForge API
// @version 1.0
// @description This is a file upload and management API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host uploader.nativespeak.app
// @BasePath /
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a JWT token.

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name Authorization
// @description API key for external services.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/database"
	_ "github.com/GoogleCloudPlatform/golang-samples/run/helloworld/docs"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/handlers"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/middleware"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/util"
	swagger "github.com/swaggo/http-swagger"
)

var (
	DB *gorm.DB
)

func main() {
	// Carrega configuração
	config.LoadConfig()

	// Conecta ao banco de dados
	var err error
	DB, err = database.Connect()
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados:", err)
	}
	log.Println("Conexão com o banco de dados estabelecida.")

	// Cria pasta uploads se não existir
	os.MkdirAll("./uploads", os.ModePerm)

	mux := http.NewServeMux()

	// Swagger UI
	mux.HandleFunc("/swagger/", swagger.WrapHandler)

	// Endpoints de autenticação (públicos)
	mux.HandleFunc("/register", handlers.RegisterHandler(DB))
	mux.HandleFunc("/login", handlers.LoginHandler(DB))
	// API Endpoints (protegidos)
	api := http.NewServeMux()
	api.HandleFunc("/upload", handlers.UploadHandler(DB))
	api.HandleFunc("/projects", handlers.ProjectsHandler(DB))
	api.HandleFunc("/list", handlers.ListHandler(DB))
	api.HandleFunc("/delete", handlers.DeleteHandler(DB))
	api.HandleFunc("/project/delete", handlers.DeleteProjectHandler(DB))
	api.HandleFunc("/user/rotate-api-key", handlers.RotateAPIKeyHandler(DB))

	// Aplica middleware de autenticação à API
	protectedAPI := middleware.AuthMiddleware(DB, api)
	mux.Handle("/api/", http.StripPrefix("/api", protectedAPI))

	// Servidor de arquivos estáticos
	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/files/", http.StripPrefix("/files/", fs))

	// Aplica o middleware de logging a todas as rotas
	loggedMux := middleware.LoggingMiddleware(mux)

	// Aplica o middleware de CORS
	corsMux := middleware.CORSMiddleware(loggedMux)

	fmt.Printf("🚀 Servidor rodando na porta %s\n", config.AppConfig.Port)
	fmt.Printf("🗄️ Database: %s\n", util.MaskDBURL(config.AppConfig.DatabaseURL))
	fmt.Printf("✅ Endpoints de API protegidos em /api/\n")
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, corsMux))
}
