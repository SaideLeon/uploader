package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/middleware"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// --- Structs para Respostas ---

type UploadResponse struct {
	Message string `json:"message"`
	URL     string `json:"url"`
	Project string `json:"project"`
	File    string `json:"file"`
}

type FileInfo struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type ListResponse struct {
	Project    string     `json:"project"`
	Files      []FileInfo `json:"files"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

type ProjectInfo struct {
	Name      string `json:"name"`
	FileCount int64  `json:"file_count"`
	TotalSize int64  `json:"total_size"`
}

type ProjectsResponse struct {
	Projects   []ProjectInfo `json:"projects"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

// --- Funções Utilitárias ---

func sanitizeProjectName(project string) string {
	project = strings.TrimSpace(project)
	project = strings.ToLower(project)
	project = strings.ReplaceAll(project, "..", "")
	project = strings.ReplaceAll(project, "/", "-")
	project = strings.ReplaceAll(project, "\\", "-")
	if project == "" {
		project = "default"
	}
	return project
}

func getPaginationParams(r *http.Request) (page, perPage int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	perPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage <= 0 || perPage > 100 {
		perPage = 10
	}
	return page, perPage
}

func calculateTotalPages(total int64, perPage int) int {
	if perPage == 0 {
		return 0
	}
	pages := int(total) / perPage
	if int(total)%perPage != 0 {
		pages++
	}
	return pages
}

const (
	MaxUploadSize = 10 * 1024 * 1024 // 10 MB
)

var (
	AllowedMimeTypes = map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}
)

// --- Handlers ---

// UploadHandler godoc
// @Summary Upload a file to a project
// @Description Uploads a file to a specified project. If the project doesn't exist, it will be created.
// @Tags api
// @Accept  multipart/form-data
// @Produce  json
// @Param   project  formData  string  true  "Project name"
// @Param   file     formData  file    true  "File to upload"
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 201 {object} UploadResponse "File uploaded successfully"
// @Failure 400 {string} string "Bad Request: Error reading file or file is too large"
// @Failure 403 {string} string "Storage limit exceeded"
// @Failure 413 {string} string "File is too large. Max size is 10MB."
// @Failure 415 {string} string "Invalid file type. Allowed types are: jpeg, png, pdf."
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/upload [post]
func UploadHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userFromCtx, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
			return
		}

		// Carrega o usuário com o plano para verificar o limite de armazenamento
		var user models.User
		if err := db.Preload("Plan").First(&user, userFromCtx.ID).Error; err != nil {
			http.Error(w, "Could not retrieve user details", http.StatusInternalServerError)
			return
		}

		// Limita o tamanho do corpo da requisição
		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
		if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
			http.Error(w, "File is too large. Max size is 10MB.", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error reading file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Verificar limite de armazenamento
		if user.StorageUsage+header.Size > user.Plan.StorageLimit {
			http.Error(w, "Storage limit exceeded", http.StatusForbidden)
			return
		}

		project_name := r.FormValue("project")
		project_name = sanitizeProjectName(project_name)

		// Valida o MIME type
		mimeType := header.Header.Get("Content-Type")
		if !AllowedMimeTypes[mimeType] {
			http.Error(w, "Invalid file type. Allowed types are: jpeg, png, pdf.", http.StatusUnsupportedMediaType)
			return
		}

		var project models.Project
		if err := db.FirstOrCreate(&project, models.Project{Name: project_name, UserID: user.ID}).Error; err != nil {
			http.Error(w, "Could not find or create project: "+err.Error(), http.StatusInternalServerError)
			return
		}

		timestamp := time.Now().Format("20060102-150405")
		ext := filepath.Ext(header.Filename)
		name := strings.TrimSuffix(header.Filename, ext)
		safeName := fmt.Sprintf("%s-%s%s", name, timestamp, ext)

		userDir := filepath.Join("./uploads", fmt.Sprintf("user_%s", user.ID.String()))
		projectDir := filepath.Join(userDir, project.Name)
		if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
			http.Error(w, "Could not create project directory: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filePath := filepath.Join(projectDir, safeName)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Could not save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		size, err := io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Error saving file content: "+err.Error(), http.StatusInternalServerError)
			return
		}

		dbFile := models.File{
			Name:      safeName,
			Path:      filePath,
			Size:      size,
			MimeType:  header.Header.Get("Content-Type"),
			ProjectID: project.ID,
		}
		if err := db.Create(&dbFile).Error; err != nil {
			http.Error(w, "Could not save file metadata: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Atualiza o uso de armazenamento do usuário
		newStorageUsage := user.StorageUsage + size
		if err := db.Model(&user).Update("storage_usage", newStorageUsage).Error; err != nil {
			// Logar o erro, mas continuar. A funcionalidade principal (upload) foi concluída.
			fmt.Printf("Error updating user storage usage for user %s: %v\n", user.ID, err)
		}

		publicURL := fmt.Sprintf("%s/files/user_%s/%s/%s", config.AppConfig.Domain, user.ID.String(), project.Name, safeName)
		resp := UploadResponse{
			Message: "File uploaded successfully",
			URL:     publicURL,
			Project: project.Name,
			File:    safeName,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

// ProjectsHandler godoc
// @Summary List user's projects
// @Description Retrieves a paginated list of projects for the authenticated user.
// @Tags api
// @Produce  json
// @Param   page      query  int  false  "Page number for pagination"
// @Param   per_page  query  int  false  "Number of items per page"
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 200 {object} ProjectsResponse
// @Router /api/projects [get]
func ProjectsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(middleware.UserContextKey).(*models.User)
		page, perPage := getPaginationParams(r)
		offset := (page - 1) * perPage

		var projects []models.Project
		db.Where("user_id = ?", user.ID).Limit(perPage).Offset(offset).Find(&projects)

		// Inicializa como slice vazio em vez de nil
		projectInfos := make([]ProjectInfo, 0)

		for _, p := range projects {
			var fileCount int64
			var totalSize int64
			db.Model(&models.File{}).Where("project_id = ?", p.ID).Count(&fileCount)
			db.Model(&models.File{}).Select("sum(size)").Where("project_id = ?", p.ID).Row().Scan(&totalSize)

			projectInfos = append(projectInfos, ProjectInfo{
				Name:      p.Name,
				FileCount: fileCount,
				TotalSize: totalSize,
			})
		}

		var totalProjects int64
		db.Model(&models.Project{}).Where("user_id = ?", user.ID).Count(&totalProjects)

		totalPages := calculateTotalPages(totalProjects, perPage)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectsResponse{
			Projects:   projectInfos,
			Total:      int(totalProjects),
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		})
	}
}

// ListHandler godoc
// @Summary List files in a project
// @Description Retrieves a paginated list of files within a specified project for the authenticated user.
// @Tags api
// @Produce  json
// @Param   project   query  string  true  "Project name"
// @Param   page      query  int     false "Page number for pagination"
// @Param   per_page  query  int     false "Number of items per page"
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 200 {object} ListResponse
// @Failure 400 {string} string "Project name is required"
// @Failure 404 {string} string "Project not found"
// @Router /api/list [get]
func ListHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(middleware.UserContextKey).(*models.User)
		projectName := r.URL.Query().Get("project")
		if projectName == "" {
			http.Error(w, "Project name is required", http.StatusBadRequest)
			return
		}

		page, perPage := getPaginationParams(r)
		offset := (page - 1) * perPage

		var project models.Project
		if err := db.First(&project, "name = ? AND user_id = ?", projectName, user.ID).Error; err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		var files []models.File
		db.Where("project_id = ?", project.ID).Limit(perPage).Offset(offset).Find(&files)

		// Inicializa como slice vazio em vez de nil
		fileInfos := make([]FileInfo, 0)
		domain := config.AppConfig.Domain

		for _, f := range files {
			fileInfos = append(fileInfos, FileInfo{
				Name:       f.Name,
				URL:        fmt.Sprintf("%s/files/user_%s/%s/%s", domain, user.ID.String(), projectName, f.Name),
				Size:       f.Size,
				UploadedAt: f.UploadedAt,
			})
		}

		var totalFiles int64
		db.Model(&models.File{}).Where("project_id = ?", project.ID).Count(&totalFiles)

		totalPages := calculateTotalPages(totalFiles, perPage)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListResponse{
			Project:    projectName,
			Files:      fileInfos,
			Total:      totalFiles,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		})
	}
}

// DeleteHandler godoc
// @Summary Delete a file
// @Description Deletes a specific file from a project.
// @Tags api
// @Produce  json
// @Param   project  query  string  true  "Project name"
// @Param   file     query  string  true  "File name"
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 200 {object} map[string]string "message: File deleted successfully"
// @Failure 400 {string} string "'project' and 'file' parameters are required"
// @Failure 404 {string} string "Project not found or File not found"
// @Failure 500 {string} string "Could not delete file metadata"
// @Router /api/delete [delete]
func DeleteHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userFromCtx, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
			return
		}

		// Carrega o usuário completo para obter o uso de armazenamento atual
		var user models.User
		if err := db.First(&user, userFromCtx.ID).Error; err != nil {
			http.Error(w, "Could not retrieve user details", http.StatusInternalServerError)
			return
		}

		projectName := r.URL.Query().Get("project")
		fileName := r.URL.Query().Get("file")

		if projectName == "" || fileName == "" {
			http.Error(w, "'project' and 'file' parameters are required", http.StatusBadRequest)
			return
		}

		var project models.Project
		if err := db.First(&project, "name = ? AND user_id = ?", projectName, user.ID).Error; err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		var file models.File
		if err := db.First(&file, "name = ? AND project_id = ?", fileName, project.ID).Error; err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Deleta o arquivo do filesystem
		if err := os.Remove(file.Path); err != nil {
			// Logar o erro, mas continuar para remover do DB
			fmt.Printf("Could not delete file from filesystem: %s\n", err.Error())
		}

		// Deleta o registro do banco de dados
		if err := db.Delete(&file).Error; err != nil {
			http.Error(w, "Could not delete file metadata: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Atualiza o uso de armazenamento do usuário
		newStorageUsage := user.StorageUsage - file.Size
		if newStorageUsage < 0 {
			newStorageUsage = 0
		}
		if err := db.Model(&user).Update("storage_usage", newStorageUsage).Error; err != nil {
			fmt.Printf("Error updating user storage usage for user %s: %v\n", user.ID, err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "File deleted successfully",
			"project": projectName,
			"file":    fileName,
		})
	}
}

// RotateAPIKeyHandler godoc
// @Summary Rotate user's API key
// @Description Generates a new API key for the authenticated user, invalidating the old one.
// @Tags api
// @Produce  json
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 200 {object} map[string]string "message: API key rotated successfully, new_api_key: ..."
// @Failure 500 {string} string "Could not rotate API key"
// @Router /api/user/rotate-api-key [post]
func RotateAPIKeyHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
			return
		}

		newAPIKey := uuid.New().String()
		if err := db.Model(&user).Update("forge_api_key", newAPIKey).Error; err != nil {
			http.Error(w, "Could not rotate API key: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":     "API key rotated successfully",
			"new_api_key": newAPIKey,
		})
	}
}

// DeleteProjectHandler godoc
// @Summary Delete an empty project
// @Description Deletes a project that has no files. Projects with files cannot be deleted.
// @Tags api
// @Produce  json
// @Param   project  query  string  true  "Project name"
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 200 {object} map[string]string "message: Project deleted successfully"
// @Failure 400 {string} string "Project name is required or Project has files and cannot be deleted"
// @Failure 404 {string} string "Project not found"
// @Failure 500 {string} string "Could not delete project"
// @Router /api/project/delete [delete]
func DeleteProjectHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(middleware.UserContextKey).(*models.User)
		projectName := r.URL.Query().Get("project")

		if projectName == "" {
			http.Error(w, "Project name is required", http.StatusBadRequest)
			return
		}

		var project models.Project
		if err := db.First(&project, "name = ? AND user_id = ?", projectName, user.ID).Error; err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		// Verificar se o projeto tem arquivos
		var fileCount int64
		db.Model(&models.File{}).Where("project_id = ?", project.ID).Count(&fileCount)

		if fileCount > 0 {
			http.Error(w, "Project has files and cannot be deleted. Delete all files first.", http.StatusBadRequest)
			return
		}

		// Deletar o projeto
		if err := db.Delete(&project).Error; err != nil {
			http.Error(w, "Could not delete project: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Tentar remover o diretório vazio
		projectDir := filepath.Join("./uploads", fmt.Sprintf("user_%s", user.ID.String()), projectName)
		os.Remove(projectDir) // Ignora erro se não existir

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Project deleted successfully",
			"project": projectName,
		})
	}
}

// UserStatusHandler godoc
// @Summary Get user status
// @Description Retrieves the current user's information, including plan and storage usage.
// @Tags api
// @Produce  json
// @Security BearerAuth
// @Security APIKeyAuth
// @Success 200 {object} models.User "User status"
// @Failure 500 {string} string "Could not retrieve user details"
// @Router /api/user/status [get]
func UserStatusHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userFromCtx, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
			return
		}

		var user models.User
		if err := db.Preload("Plan").First(&user, userFromCtx.ID).Error; err != nil {
			http.Error(w, "Could not retrieve user details", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
