package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	FreePlanStorageLimit = 1 * 1024 * 1024 * 1024 // 1 GB
)

// Plan representa um plano de assinatura
type Plan struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name         string    `gorm:"uniqueIndex;not null"`
	Price        float64   `gorm:"not null;default:0"`
	StorageLimit int64     `gorm:"not null"` // Em bytes
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// User representa um usuário no sistema
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	Password     string    `gorm:"not null"`
	ForgeAPIKey  string    `gorm:"uniqueIndex;not null"`
	StorageUsage int64     `gorm:"default:0"`
	PlanID       uuid.UUID `gorm:"type:uuid"`
	Plan         Plan      `gorm:"foreignKey:PlanID"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	Projects     []Project `gorm:"foreignKey:UserID"`
}

// Project representa um projeto de um usuário
type Project struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name      string    `gorm:"uniqueIndex:idx_user_project;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_project;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Files     []File    `gorm:"foreignKey:ProjectID"`
}

// File representa um arquivo enviado para um projeto
type File struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name       string    `gorm:"not null"`
	Path       string    `gorm:"not null"`
	Size       int64     `gorm:"not null"`
	MimeType   string    `gorm:"not null"`
	ProjectID  uuid.UUID `gorm:"type:uuid;not null"`
	UploadedAt time.Time `gorm:"autoCreateTime"`
}

// BeforeCreate é um hook do GORM para gerar um UUID para o plano
func (p *Plan) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}

// BeforeCreate é um hook do GORM para gerar a API key e hashear a senha antes de criar um usuário
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	// Gerar API Key
	u.ForgeAPIKey = uuid.New().String()

	// Hashear a senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	return nil
}

// BeforeCreate is a GORM hook to generate a UUID for the project ID before creating a record
func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}

// BeforeCreate is a GORM hook to generate a UUID for the file ID before creating a record
func (f *File) BeforeCreate(tx *gorm.DB) (err error) {
	f.ID = uuid.New()
	return
}

// CheckPassword compara a senha fornecida com o hash armazenado
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
