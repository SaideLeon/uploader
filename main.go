		package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Struct para retorno JSON do upload
type UploadResponse struct {
	Message string `json:"message"`
	URL     string `json:"url"`
	Project string `json:"project"`
	File    string `json:"file"`
}

// Struct para listagem de arquivos
type FileInfo struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	UploadedAt string   `json:"uploaded_at"`
}

type ListResponse struct {
	Project string     `json:"project"`
	Files   []FileInfo `json:"files"`
	Total   int        `json:"total"`
}

// Struct para listagem de projetos
type ProjectInfo struct {
	Name       string `json:"name"`
	FileCount  int    `json:"file_count"`
	TotalSize  int64  `json:"total_size"`
}

type ProjectsResponse struct {
	Projects []ProjectInfo `json:"projects"`
	Total    int           `json:"total"`
}

// Sanitiza o nome do projeto para evitar problemas
func sanitizeProjectName(project string) string {
	// Remove espaços e caracteres especiais
	project = strings.TrimSpace(project)
	project = strings.ToLower(project)
	// Remove caracteres perigosos
	project = strings.ReplaceAll(project, "..", "")
	project = strings.ReplaceAll(project, "/", "-")
	project = strings.ReplaceAll(project, "\\", "-")
	
	if project == "" {
		project = "default"
	}
	
	return project
}

// Obtém o domínio baseado no ambiente
func getDomain() string {
	env := os.Getenv("ENV")
	domain := os.Getenv("DOMAIN_LOCAL")
	if env == "production" {
		domain = os.Getenv("DOMAIN_PROD")
	}
	return domain
}

// UploadHandler com suporte a projetos
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Use POST para enviar arquivos",
		})
		return
	}

	// Obtém o nome do projeto (padrão: "default")
	project := r.FormValue("project")
	project = sanitizeProjectName(project)

	// Lê o arquivo
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao ler arquivo: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Gera timestamp formatado
	timestamp := time.Now().Format("20060102-150405")

	// Separa nome e extensão
	ext := filepath.Ext(header.Filename)
	name := header.Filename[:len(header.Filename)-len(ext)]

	// Nome único
	newFileName := fmt.Sprintf("%s-%s%s", name, timestamp, ext)

	// Cria diretório do projeto se não existir
	projectDir := filepath.Join("./uploads", project)
	if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao criar diretório do projeto: " + err.Error(),
		})
		return
	}

	// Cria arquivo no diretório do projeto
	dstPath := filepath.Join(projectDir, newFileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao salvar arquivo: " + err.Error(),
		})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao copiar arquivo: " + err.Error(),
		})
		return
	}

	// Gera URL pública
	domain := getDomain()
	publicURL := fmt.Sprintf("%s/files/%s/%s", domain, project, newFileName)

	// Retorna JSON
	resp := UploadResponse{
		Message: "Arquivo enviado com sucesso",
		URL:     publicURL,
		Project: project,
		File:    newFileName,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListHandler lista arquivos de um projeto específico
func listHandler(w http.ResponseWriter, r *http.Request) {
	// Obtém o projeto da query string
	project := r.URL.Query().Get("project")
	project = sanitizeProjectName(project)

	projectDir := filepath.Join("./uploads", project)

	// Verifica se o diretório existe
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListResponse{
			Project: project,
			Files:   []FileInfo{},
			Total:   0,
		})
		return
	}

	files, err := os.ReadDir(projectDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao ler arquivos: " + err.Error(),
		})
		return
	}

	domain := getDomain()
	fileInfos := []FileInfo{}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		info, err := f.Info()
		if err != nil {
			continue
		}

		fileInfos = append(fileInfos, FileInfo{
			Name:      f.Name(),
			URL:       fmt.Sprintf("%s/files/%s/%s", domain, project, f.Name()),
			Size:      info.Size(),
			UploadedAt: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	resp := ListResponse{
		Project: project,
		Files:   fileInfos,
		Total:   len(fileInfos),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ProjectsHandler lista todos os projetos disponíveis
func projectsHandler(w http.ResponseWriter, r *http.Request) {
	uploadsDir := "./uploads"

	entries, err := os.ReadDir(uploadsDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao ler diretório: " + err.Error(),
		})
		return
	}

	projects := []ProjectInfo{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(uploadsDir, entry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}

		var totalSize int64
		fileCount := 0

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			info, err := file.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
			fileCount++
		}

		projects = append(projects, ProjectInfo{
			Name:      entry.Name(),
			FileCount: fileCount,
			TotalSize: totalSize,
		})
	}

	resp := ProjectsResponse{
		Projects: projects,
		Total:    len(projects),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteHandler deleta um arquivo específico
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Use DELETE para remover arquivos",
		})
		return
	}

	project := r.URL.Query().Get("project")
	fileName := r.URL.Query().Get("file")

	if project == "" || fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Parâmetros 'project' e 'file' são obrigatórios",
		})
		return
	}

	project = sanitizeProjectName(project)
	filePath := filepath.Join("./uploads", project, fileName)

	if err := os.Remove(filePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao deletar arquivo: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Arquivo deletado com sucesso",
		"project": project,
		"file":    fileName,
	})
}

func main() {
	// Carrega variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Não foi possível carregar .env, usando variáveis do sistema")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	// Cria pasta uploads se não existir
	os.MkdirAll("./uploads", os.ModePerm)

	// Endpoints
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/projects", projectsHandler)
	http.HandleFunc("/delete", deleteHandler)
	
	// Servidor de arquivos estáticos
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	domain := getDomain()

	fmt.Printf("🚀 Servidor rodando na porta %s\n", port)
	fmt.Printf("📤 Upload: %s/upload (POST com 'file' e 'project')\n", domain)
	fmt.Printf("📋 Listar projetos: %s/projects\n", domain)
	fmt.Printf("📁 Listar arquivos: %s/list?project={nome}\n", domain)
	fmt.Printf("📥 Download: %s/files/{projeto}/{arquivo}\n", domain)
	fmt.Printf("🗑️  Deletar: %s/delete?project={nome}&file={arquivo} (DELETE)\n", domain)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
