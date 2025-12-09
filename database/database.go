package database

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// Connect abre a conexão com o banco de dados e executa as migrações
func Connect() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.AppConfig.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	log.Println("Executando migrações do banco de dados...")
	err = db.AutoMigrate(&models.User{}, &models.Project{}, &models.File{})
	if err != nil {
		log.Println("Erro ao executar migrações:", err)
		return nil, err
	}

	return db, nil
}
