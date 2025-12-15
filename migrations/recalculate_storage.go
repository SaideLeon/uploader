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
	config.LoadConfig()
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	RecalculateStorageUsage(db)
}

func RecalculateStorageUsage(db *gorm.DB) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Error fetching users: %v", err)
	}

	fmt.Printf("Found %d users to process.\n", len(users))

	for _, user := range users {
		var totalSize int64
		// Subquery to get all project IDs for the user
		projectIDs := db.Model(&models.Project{}).Select("id").Where("user_id = ?", user.ID)

		// Sum the size of all files belonging to those projects
		if err := db.Model(&models.File{}).Select("sum(size)").Where("project_id IN (?)", projectIDs).Row().Scan(&totalSize); err != nil {
			log.Printf("Could not calculate storage for user %s: %v", user.Email, err)
			continue
		}

		if user.StorageUsage != totalSize {
			fmt.Printf("Updating user %s: Old storage %d, New storage %d\n", user.Email, user.StorageUsage, totalSize)
			if err := db.Model(&user).Update("storage_usage", totalSize).Error; err != nil {
				log.Printf("Failed to update storage for user %s: %v", user.Email, err)
			}
		} else {
			fmt.Printf("User %s storage is already up to date.\n", user.Email)
		}
	}

	fmt.Println("Storage recalculation complete.")
}
