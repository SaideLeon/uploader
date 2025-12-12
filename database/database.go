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

// ConnectTest abre uma conexão com um banco de dados SQLite em memória para testes
func ConnectTest() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.User{}, &models.Project{}, &models.File{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CloseTest fecha a conexão com o banco de dados de teste
func CloseTest(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Falha ao obter o objeto sql.DB:", err)
	}
	sqlDB.Close()
}
