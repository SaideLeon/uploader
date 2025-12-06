# 📦 Upload Service - Organização por Projetos

Serviço de upload de arquivos com organização automática por projetos/diretórios.

## 🚀 Iniciar o Servidor

```bash
go run main.go
```

## 📁 Estrutura de Diretórios

```
uploads/
├── projeto-a/
│   ├── imagem1-20240101-120000.jpg
│   └── imagem2-20240101-120100.jpg
├── projeto-b/
│   └── documento-20240101-120200.pdf
└── default/
    └── arquivo-20240101-120300.txt
```

## 🔌 Endpoints Disponíveis

### 1. Upload de Arquivo
**POST** `/upload`

Faz upload de um arquivo para um projeto específico.

**Parâmetros:**
- `file` (form-data, obrigatório): O arquivo a ser enviado
- `project` (form-data, opcional): Nome do projeto (padrão: "default")

**Exemplo com cURL:**
```bash
curl -X POST http://localhost:8002/upload \
  -F "file=@/caminho/para/imagem.jpg" \
  -F "project=meu-app"
```

**Exemplo com JavaScript:**
```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);
formData.append('project', 'meu-app');

fetch('http://localhost:8002/upload', {
  method: 'POST',
  body: formData
})
.then(res => res.json())
.then(data => console.log(data));
```

**Resposta:**
```json
{
  "message": "Arquivo enviado com sucesso",
  "url": "http://localhost:8002/files/meu-app/imagem-20240101-120000.jpg",
  "project": "meu-app",
  "file": "imagem-20240101-120000.jpg"
}
```

---

### 2. Listar Todos os Projetos
**GET** `/projects`

Lista todos os projetos disponíveis com estatísticas.

**Exemplo:**
```bash
curl http://localhost:8002/projects
```

**Resposta:**
```json
{
  "projects": [
    {
      "name": "meu-app",
      "file_count": 15,
      "total_size": 2048576
    },
    {
      "name": "outro-projeto",
      "file_count": 8,
      "total_size": 1024000
    }
  ],
  "total": 2
}
```

---

### 3. Listar Arquivos de um Projeto
**GET** `/list?project={nome}`

Lista todos os arquivos de um projeto específico.

**Exemplo:**
```bash
curl http://localhost:8002/list?project=meu-app
```

**Resposta:**
```json
{
  "project": "meu-app",
  "files": [
    {
      "name": "imagem-20240101-120000.jpg",
      "url": "http://localhost:8002/files/meu-app/imagem-20240101-120000.jpg",
      "size": 204800,
      "uploaded_at": "2024-01-01 12:00:00"
    }
  ],
  "total": 1
}
```

---

### 4. Acessar/Baixar Arquivo
**GET** `/files/{projeto}/{arquivo}`

Acessa ou baixa um arquivo específico.

**Exemplo:**
```bash
curl http://localhost:8002/files/meu-app/imagem-20240101-120000.jpg -o imagem.jpg
```

---

### 5. Deletar Arquivo
**DELETE** `/delete?project={nome}&file={arquivo}`

Remove um arquivo específico de um projeto.

**Exemplo:**
```bash
curl -X DELETE "http://localhost:8002/delete?project=meu-app&file=imagem-20240101-120000.jpg"
```

**Resposta:**
```json
{
  "message": "Arquivo deletado com sucesso",
  "project": "meu-app",
  "file": "imagem-20240101-120000.jpg"
}
```

## 🎯 Casos de Uso

### Exemplo 1: Upload de Imagens de um App Mobile
```javascript
// No seu app
const uploadImage = async (imageFile, appName) => {
  const formData = new FormData();
  formData.append('file', imageFile);
  formData.append('project', appName);
  
  const response = await fetch('https://uploader.nativespeak.app/upload', {
    method: 'POST',
    body: formData
  });
  
  const data = await response.json();
  return data.url; // Use esta URL no seu app
};
```

### Exemplo 2: Filtrar Imagens por Projeto
```javascript
// Listar apenas imagens do projeto "app-vendas"
const response = await fetch('http://localhost:8002/list?project=app-vendas');
const data = await response.json();

data.files.forEach(file => {
  console.log(`${file.name} - ${file.size} bytes`);
});
```

### Exemplo 3: Galeria de Imagens por Projeto
```javascript
// Criar galeria HTML
const createGallery = async (projectName) => {
  const response = await fetch(`http://localhost:8002/list?project=${projectName}`);
  const data = await response.json();
  
  const gallery = document.getElementById('gallery');
  data.files.forEach(file => {
    const img = document.createElement('img');
    img.src = file.url;
    img.alt = file.name;
    gallery.appendChild(img);
  });
};
```

## ⚙️ Configuração (.env)

```env
ENV=local
DOMAIN_LOCAL=http://localhost:8002
DOMAIN_PROD=https://uploader.nativespeak.app
PORT=8002
```

## 🔒 Segurança

- Nomes de projetos são sanitizados automaticamente
- Caracteres perigosos (`..`, `/`, `\`) são removidos
- Cada projeto tem seu próprio diretório isolado

## 📝 Notas

- Arquivos recebem timestamp automático para evitar conflitos
- Se nenhum projeto for especificado, usa "default"
- Projetos são criados automaticamente no primeiro upload
- Nomes de projeto são convertidos para lowercase

## 🛠️ Tecnologias

- Go 1.20+
- Pacote `godotenv` para variáveis de ambiente
