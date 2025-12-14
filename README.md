# 🚀 Forge Uploader - API Segura de Upload de Arquivos

O Forge Uploader é um serviço de upload de arquivos robusto e seguro, construído em Go. Ele oferece autenticação de usuários, gerenciamento de chaves de API, organização de arquivos por projetos e políticas de segurança avançadas.

## ✨ Features
- **Autenticação de Usuários**: Sistema de contas com E-mail/Senha e autenticação baseada em JWT.
- **Chave de API**: Cada usuário recebe uma `FORGE_API_KEY` para autenticar requisições.
- **Namespace por Usuário**: Cada usuário tem seu próprio escopo de projetos, garantindo isolamento e segurança.
- **Políticas de Segurança**:
  - Limite de tamanho de arquivo (10MB por upload).
  - Validação de Mime-Type (`image/jpeg`, `image/png`, `application/pdf`).
  - Rate Limiting (100 uploads por dia por usuário).
  - Logs de auditoria para todas as requisições.
- **Paginação**: Endpoints de listagem (`/api/projects`, `/api/list`) são paginados.
- **Armazenamento Flexível**: Estrutura preparada para futuros drivers (S3, MinIO, etc.).

## 🚀 Iniciar o Servidor

1.  **Configure o `.env`**:
    Copie o `.env.example` para `.env` e ajuste as variáveis, se necessário.
    ```env
    # Ambiente: "local" ou "production"
    ENV=local

    # Porta do servidor
    PORT=8002

    # Chave secreta para JWT (troque por um valor seguro em produção)
    JWT_SECRET=your-super-secret-jwt-key

    # Caminho para o banco de dados SQLite
    DATABASE_URL=forge.db
    ```

2.  **Execute o servidor**:
    ```bash
    go run main.go
    ```

## 🔌 Endpoints da API

Todos os endpoints da API estão sob o prefixo `/api` e exigem autenticação.

**Autenticação**:
Forneça o Token JWT ou a `FORGE_API_KEY` no header `Authorization`.

```
Authorization: Bearer <SEU_TOKEN_JWT_OU_API_KEY>
```

---

### 👤 Autenticação

#### 1. Criar Conta
**POST** `/register`

Cria um novo usuário e retorna a `FORGE_API_KEY` inicial.

**Body (JSON)**:
```json
{
  "email": "user@example.com",
  "password": "your-strong-password"
}
```

#### 2. Fazer Login
**POST** `/login`

Autentica um usuário e retorna um Token JWT válido por 24 horas.

**Body (JSON)**:
```json
{
  "email": "user@example.com",
  "password": "your-strong-password"
}
```

#### 3. Rotacionar a Chave de API
**POST** `/api/user/rotate-api-key`

Gera uma nova `FORGE_API_KEY` para o usuário autenticado.

---

### 📦 Arquivos e Projetos

#### 1. Upload de Arquivo
**POST** `/api/upload`

Faz upload de um arquivo para um projeto. Se o projeto não existir, ele é criado.

**Parâmetros (form-data)**:
- `file` (obrigatório): O arquivo a ser enviado.
- `project` (opcional): Nome do projeto (padrão: "default").

**Exemplo com cURL**:
```bash
curl -X POST http://localhost:8002/api/upload \
  -H "Authorization: Bearer <SUA_API_KEY>" \
  -F "file=@/path/to/image.png" \
  -F "project=my-app"
```

#### 2. Listar Projetos
**GET** `/api/projects`

Lista os projetos do usuário com estatísticas.

**Query Params (opcional)**:
- `page`: Número da página.
- `per_page`: Itens por página.

#### 3. Listar Arquivos de um Projeto
**GET** `/api/list?project={nome}`

Lista os arquivos de um projeto específico.

**Query Params (opcional)**:
- `page`: Número da página.
- `per_page`: Itens por página.

#### 4. Deletar Arquivo
**DELETE** `/api/delete?project={nome}&file={arquivo}`

Remove um arquivo de um projeto.

---

### 📂 Acesso a Arquivos

#### Acessar/Baixar Arquivo
**GET** `/files/{user_id}/{projeto}/{arquivo}`

Acessa um arquivo enviado. A URL é retornada na resposta do upload.

**Exemplo**:
```bash
curl http://localhost:8002/files/user_1/my-app/image-20251209-174000.png -o image.png
```

## 🛠️ Tecnologias

- Go 1.21+
- GORM (com driver SQLite CGO-free)
- JWT para autenticação
- `godotenv` para variáveis de ambiente
---
## 📚 Documentação Detalhada e Exemplos de Integração

Esta seção fornece exemplos práticos de como interagir com a **MidiaForge API** em várias linguagens de programação.

A URL base para todas as requisições da API é:
```
https://uploader.nativespeak.app/
```

### Autenticação

A autenticação pode ser feita de duas formas:
1.  **Token JWT**: Obtido no endpoint `/login`. Válido por 24 horas.
2.  **Chave de API**: Obtida no momento do registro (`/register`) ou ao rotacionar a chave (`/api/user/rotate-api-key`).

Em ambos os casos, o token ou a chave devem ser enviados no cabeçalho `Authorization`.

```
Authorization: Bearer <SEU_TOKEN_JWT_OU_API_KEY>
```

--- 

###  exemplos de código

A seguir, exemplos de como realizar as principais operações na API.

#### **cURL**

**1. Registrar um Novo Usuário**
```bash
curl -X POST https://uploader.nativespeak.app/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your-strong-password"
  }'
```
> **Resposta Esperada**: Um JSON contendo `token` e `forge_api_key`. Guarde a chave de API em um local seguro.

**2. Fazer Login**
```bash
curl -X POST https://uploader.nativespeak.app/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your-strong-password"
  }'
```
> **Resposta Esperada**: Um JSON com um novo `token` JWT.

**3. Fazer Upload de um Arquivo**
```bash
curl -X POST https://uploader.nativespeak.app/api/upload \
  -H "Authorization: Bearer <SUA_API_KEY_OU_TOKEN>" \
  -F "file=@/path/to/your/file.png" \
  -F "project=my-awesome-project"
```
> **Resposta Esperada**: Um JSON com a URL do arquivo (`url`), nome do arquivo (`file`), e projeto (`project`).

**4. Listar Projetos**
```bash
curl -X GET "https://uploader.nativespeak.app/api/projects?page=1&per_page=10" \
  -H "Authorization: Bearer <SUA_API_KEY_OU_TOKEN>"
```

**5. Listar Arquivos de um Projeto**
```bash
curl -X GET "https://uploader.nativespeak.app/api/list?project=my-awesome-project" \
  -H "Authorization: Bearer <SUA_API_KEY_OU_TOKEN>"
```

**6. Deletar um Arquivo**
```bash
curl -X DELETE "https://uploader.nativespeak.app/api/delete?project=my-awesome-project&file=file.png" \
  -H "Authorization: Bearer <SUA_API_KEY_OU_TOKEN>"
```

--- 

#### **JavaScript/TypeScript (com `fetch`)**

```typescript
const API_BASE_URL = 'https://uploader.nativespeak.app';
const API_KEY = 'SUA_API_KEY_OU_TOKEN';

// 1. Registrar um Novo Usuário
async function registerUser(email, password) {
  const response = await fetch(`${API_BASE_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  return response.json();
}

// 2. Fazer Login
async function login(email, password) {
  const response = await fetch(`${API_BASE_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  return response.json();
}

// 3. Fazer Upload de um Arquivo
async function uploadFile(file: File, project: string) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('project', project);

  const response = await fetch(`${API_BASE_URL}/api/upload`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
    },
    body: formData,
  });
  return response.json();
}

// Exemplo de uso do upload
const fileInput = document.querySelector<HTMLInputElement>('input[type="file"]');
if (fileInput && fileInput.files.length > 0) {
  uploadFile(fileInput.files[0], 'my-frontend-project').then(data => {
    console.log('Arquivo enviado:', data.url);
  });
}

// 4. Listar Projetos
async function listProjects() {
    const response = await fetch(`${API_BASE_URL}/api/projects`, {
        headers: { 'Authorization': `Bearer ${API_KEY}` }
    });
    return response.json();
}
```

--- 

#### **ReactJS (Exemplo de componente de Upload)**

```jsx
import React, { useState } from 'react';

const API_BASE_URL = 'https://uploader.nativespeak.app';
const API_KEY = 'SUA_API_KEY_OU_TOKEN';

function FileUploader() {
  const [file, setFile] = useState<File | null>(null);
  const [project, setProject] = useState('react-project');
  const [uploading, setUploading] = useState(false);
  const [uploadUrl, setUploadUrl] = useState('');

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files) {
      setFile(event.target.files[0]);
    }
  };

  const handleUpload = async () => {
    if (!file) {
      alert('Por favor, selecione um arquivo.');
      return;
    }

    setUploading(true);
    setUploadUrl('');

    const formData = new FormData();
    formData.append('file', file);
    formData.append('project', project);

    try {
      const response = await fetch(`${API_BASE_URL}/api/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${API_KEY}`,
        },
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Falha no upload: ' + await response.text());
      }

      const data = await response.json();
      setUploadUrl(data.url);
      alert(`Upload bem-sucedido! URL: ${data.url}`);
    } catch (error) {
      console.error(error);
      alert(error.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input type="text" value={project} onChange={(e) => setProject(e.target.value)} placeholder="Nome do Projeto" />
      <input type="file" onChange={handleFileChange} />
      <button onClick={handleUpload} disabled={uploading}>
        {uploading ? 'Enviando...' : 'Fazer Upload'}
      </button>
      {uploadUrl && (
        <p>Arquivo enviado: <a href={uploadUrl} target="_blank" rel="noopener noreferrer">{uploadUrl}</a></p>
      )}
    </div>
  );
}

export default FileUploader;
```

--- 

#### **Python (com `requests`)**

```python
import requests
import json

API_BASE_URL = "https://uploader.nativespeak.app"
API_KEY = "SUA_API_KEY_OU_TOKEN"

# 1. Registrar um Novo Usuário
def register_user(email, password):
    url = f"{API_BASE_URL}/register"
    payload = {"email": email, "password": password}
    response = requests.post(url, json=payload)
    return response.json()

# 2. Fazer Login
def login(email, password):
    url = f"{API_BASE_URL}/login"
    payload = {"email": email, "password": password}
    response = requests.post(url, json=payload)
    return response.json()

# 3. Fazer Upload de um Arquivo
def upload_file(file_path, project_name="python-project"):
    url = f"{API_BASE_URL}/api/upload"
    headers = {"Authorization": f"Bearer {API_KEY}"}
    files = {'file': open(file_path, 'rb')}
    data = {'project': project_name}
    
    response = requests.post(url, headers=headers, files=files, data=data)
    return response.json()

# Exemplo de uso
# api_data = register_user("test@example.com", "securepass123")
# print("Chave de API:", api_data.get("forge_api_key"))

# upload_response = upload_file("path/to/your/image.jpg", "my-py-project")
# print("Arquivo enviado:", upload_response.get("url"))

# 4. Listar Projetos
def list_projects():
    url = f"{API_BASE_URL}/api/projects"
    headers = {"Authorization": f"Bearer {API_KEY}"}
    response = requests.get(url, headers=headers)
    return response.json()

# print(list_projects())
```

--- 

#### **Go (com `net/http`)**

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	apiBaseURL = "https://uploader.nativespeak.app"
	apiKey     = "SUA_API_KEY_OU_TOKEN"
)

// UploadFile faz upload de um arquivo para a API
func UploadFile(filePath, projectName string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Adiciona o campo 'project'
	_ = writer.WriteField("project", projectName)

	// Adiciona o arquivo
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/upload", apiBaseURL), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	// Exemplo de uso
	response, err := UploadFile("path/to/your/file.pdf", "my-go-project")
	if err != nil {
		fmt.Printf("Erro no upload: %v\n", err)
		return
	}
	fmt.Printf("Upload bem-sucedido! URL: %s\n", response["url"])
}
```

--- 

#### **Java (com `OkHttp`)**

Para usar este exemplo, adicione a dependência do OkHttp ao seu projeto (Maven/Gradle).

```xml
<!-- Exemplo com Maven -->
<dependency>
    <groupId>com.squareup.okhttp3</groupId>
    <artifactId>okhttp</artifactId>
    <version>4.9.3</version>
</dependency>
```

```java
import okhttp3.*;
import java.io.File;
import java.io.IOException;

public class ForgeUploaderClient {

    private static final String API_BASE_URL = "https://uploader.nativespeak.app";
    private static final String API_KEY = "SUA_API_KEY_OU_TOKEN";
    private static final OkHttpClient client = new OkHttpClient();

    public static String uploadFile(String filePath, String projectName) throws IOException {
        File file = new File(filePath);
        String fileName = file.getName();

        RequestBody requestBody = new MultipartBody.Builder()
                .setType(MultipartBody.FORM)
                .addFormDataPart("project", projectName)
                .addFormDataPart("file", fileName,
                        RequestBody.create(file, MediaType.parse("application/octet-stream")))
                .build();

        Request request = new Request.Builder()
                .url(API_BASE_URL + "/api/upload")
                .header("Authorization", "Bearer " + API_KEY)
                .post(requestBody)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Falha no upload: " + response.body().string());
            }
            return response.body().string();
        }
    }

    public static void main(String[] args) {
        try {
            String response = uploadFile("path/to/your/image.png", "my-java-project");
            System.out.println("Resposta da API: " + response);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
```