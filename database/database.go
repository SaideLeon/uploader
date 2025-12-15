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
	// Em desenvolvimento, podemos dropar as tabelas para garantir a atualização do esquema.
	// CUIDADO: Isso apagará todos os dados. Não use em produção.
	db.Migrator().DropTable(&models.User{}, &models.Project{}, &models.File{}, &models.Plan{})
	err = db.AutoMigrate(&models.User{}, &models.Project{}, &models.File{}, &models.Plan{})
	if err != nil {
		log.Println("Erro ao executar migrações:", err)
		return nil, err
	}

	// Cria o plano padrão se não existir
	CreateDefaultPlan(db)

	return db, nil
}

// CreateDefaultPlan cria o plano gratuito se ele não existir
func CreateDefaultPlan(db *gorm.DB) {
	var freePlan models.Plan
	// Verifica se o plano "Free" já existe
	if err := db.Where("name = ?", "Free").First(&freePlan).Error; err == gorm.ErrRecordNotFound {
		// Plano não encontrado, então cria um novo
		log.Println("Criando plano 'Free' padrão...")
		newFreePlan := models.Plan{
			Name:         "Free",
			Price:        0,
			StorageLimit: models.FreePlanStorageLimit, // 1 GB
		}
		if err := db.Create(&newFreePlan).Error; err != nil {
			log.Fatalf("Falha ao criar o plano 'Free': %v", err)
		}
		log.Println("Plano 'Free' criado com sucesso.")
	} else if err != nil {
		// Outro erro ocorreu durante a busca
		log.Fatalf("Erro ao verificar o plano 'Free': %v", err)
	} else {
		// Plano já existe
		log.Println("Plano 'Free' já existe.")
	}
}


// ConnectTest abre uma conexão com um banco de dados de teste
func ConnectTest() (*gorm.DB, error) {
	// ATENÇÃO: Isso usará a mesma URL de banco de dados da aplicação.
	// Para testes de verdade, considere usar um banco de dados de teste separado.
	db, err := gorm.Open(postgres.Open(config.AppConfig.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Limpa o banco de dados de teste antes de executar as migrações
	err = db.Migrator().DropTable(&models.User{}, &models.Project{}, &models.File{}, &models.Plan{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.User{}, &models.Project{}, &models.File{}, &models.Plan{})
	if err != nil {
		return nil, err
	}

	// Cria o plano padrão para os testes
	CreateDefaultPlan(db)

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
