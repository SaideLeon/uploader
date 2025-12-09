package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User representa um usuário no sistema
type User struct {
	ID          uint      `gorm:"primaryKey"`
	Email       string    `gorm:"uniqueIndex;not null"`
	Password    string    `gorm:"not null"`
	ForgeAPIKey string    `gorm:"uniqueIndex;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Projects    []Project `gorm:"foreignKey:UserID"`
}

// Project representa um projeto de um usuário
type Project struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex:idx_user_project;not null"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_project;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Files     []File    `gorm:"foreignKey:ProjectID"`
}

// File representa um arquivo enviado para um projeto
type File struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Path        string    `gorm:"not null"`
	Size        int64     `gorm:"not null"`
	MimeType    string    `gorm:"not null"`
	ProjectID   uint      `gorm:"not null"`
	UploadedAt  time.Time `gorm:"autoCreateTime"`
}

// BeforeCreate é um hook do GORM para gerar a API key e hashear a senha antes de criar um usuário
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Gerar API Key
	u.ForgeAPIKey = uuid.New().String()

	// Hashear a senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	return nil
}

// CheckPassword compara a senha fornecida com o hash armazenado
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
