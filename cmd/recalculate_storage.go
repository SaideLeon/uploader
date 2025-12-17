package main

import (
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/database"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
	"gorm.io/gorm"
)

func main() {
	// Carrega configuração
	config.LoadConfig()
	
	// Conecta ao banco
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 Starting storage usage recalculation...")
	RecalculateStorageUsage(db)
	fmt.Println("✅ Recalculation complete!")
}

// RecalculateStorageUsage recalcula o uso de armazenamento de todos os usuários
func RecalculateStorageUsage(db *gorm.DB) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("❌ Error fetching users: %v", err)
	}

	fmt.Printf("📊 Found %d users to process.\n\n", len(users))

	for _, user := range users {
		var totalSize int64
		
		// Subquery: busca todos os IDs de projetos do usuário
		projectIDs := db.Model(&models.Project{}).Select("id").Where("user_id = ?", user.ID)

		// Soma o tamanho de todos os arquivos desses projetos
		if err := db.Model(&models.File{}).
			Select("COALESCE(sum(size), 0)").
			Where("project_id IN (?)", projectIDs).
			Row().
			Scan(&totalSize); err != nil {
			log.Printf("⚠️  Warning: Could not calculate storage for user %s (%s): %v", 
				user.Email, user.ID, err)
			continue
		}

		// Verifica se precisa atualizar
		if user.StorageUsage != totalSize {
			fmt.Printf("🔧 Updating user: %s\n", user.Email)
			fmt.Printf("   Old storage: %d bytes (%.2f MB)\n", user.StorageUsage, float64(user.StorageUsage)/(1024*1024))
			fmt.Printf("   New storage: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/(1024*1024))
			
			if err := db.Model(&user).Update("storage_usage", totalSize).Error; err != nil {
				log.Printf("❌ Failed to update storage for user %s: %v\n", user.Email, err)
			} else {
				fmt.Printf("✅ Successfully updated!\n\n")
			}
		} else {
			fmt.Printf("✓ User %s storage is already correct (%d bytes)\n\n", user.Email, user.StorageUsage)
		}
	}
}
