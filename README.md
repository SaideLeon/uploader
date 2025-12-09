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