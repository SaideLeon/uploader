package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// Connect abre a conexão com o banco de dados e executa as migrações
func Connect() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.AppConfig.DatabaseURL), &gorm.Config{
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

// ConnectTest abre uma conexão com um banco de dados de teste
func ConnectTest() (*gorm.DB, error) {
	// ATENÇÃO: Isso usará a mesma URL de banco de dados da aplicação.
	// Para testes de verdade, considere usar um banco de dados de teste separado.
	db, err := gorm.Open(postgres.Open(config.AppConfig.DatabaseURL), &gorm.Config{})
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
