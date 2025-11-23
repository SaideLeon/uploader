package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// Struct para retorno JSON
type UploadResponse struct {
	Message string `json:"message"`
	URL     string `json:"url"`
}

// UploadHandler com timestamp no nome do arquivo
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Use POST para enviar arquivos")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Erro ao ler arquivo:", err)
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

	// Cria arquivo na pasta uploads
	dstPath := "./uploads/" + newFileName
	dst, err := os.Create(dstPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erro ao salvar arquivo:", err)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erro ao copiar arquivo:", err)
		return
	}

	// Define domínio correto baseado no ambiente
	env := os.Getenv("ENV")
	domain := os.Getenv("DOMAIN_LOCAL")
	if env == "production" {
		domain = os.Getenv("DOMAIN_PROD")
	}

	publicURL := fmt.Sprintf("%s/files/%s", domain, newFileName)

	// Retorna JSON
	resp := UploadResponse{
		Message: "Arquivo enviado com sucesso",
		URL:     publicURL,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListHandler lista todos os arquivos disponíveis
func listHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("./uploads")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erro ao ler arquivos:", err)
		return
	}

	fmt.Fprintln(w, "Arquivos disponíveis:")
	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
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
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	// Define domínio baseado no ambiente
	env := os.Getenv("ENV")
	domain := os.Getenv("DOMAIN_LOCAL")
	if env == "production" {
		domain = os.Getenv("DOMAIN_PROD")
	}

	fmt.Printf("Servidor rodando na porta %s\n", port)
	fmt.Printf("Upload: %s/upload\n", domain)
	fmt.Printf("Listar arquivos: %s/list\n", domain)
	fmt.Printf("Download: %s/files/{nome_do_arquivo}\n", domain)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
