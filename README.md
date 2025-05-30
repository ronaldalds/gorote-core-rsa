# Gorote Core RSA - Authentication Library

![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)
![Fiber](https://img.shields.io/badge/Fiber-2.x-00ADD8)
![JWT](https://img.shields.io/badge/JWT-RSA-000000?logo=JSON%20web%20tokens)
[![Go Report Card](https://goreportcard.com/badge/github.com/ronaldalds/gorote-core-rsa)](https://goreportcard.com/report/github.com/ronaldalds/gorote-core-rsa)

Uma biblioteca completa para autenticação e autorização usando JWT com criptografia RSA, desenvolvida em Go com Fiber.

## 📋 Pré-requisitos

- Go 1.20 ou superior
- PostgreSQL (13 ou superior)
- OpenSSL (para geração de chaves)
- Git

## 🚀 Começando

### 1. Clonar o repositório

```bash
git clone https://github.com/ronaldalds/gorote-core-rsa.git
cd gorote-core-rsa
```

### 2. Configurar ambiente

Gerar chave RSA

#### 2.1 Gerar chave privada (2048 bits)
```bash
openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048
```

#### 2.2 Extrair chave pública
```bash
openssl rsa -pubout -in private_key.pem -out public_key.pem
```

### 2. Configurar banco de dados

Crie dois bancos de dados no PostgreSQL:

```sql
-- Banco para o serviço de autenticação
CREATE DATABASE gorote;

-- Banco para microserviços
CREATE DATABASE service;
```

## 🛠️ Configuração do Projeto

Para uma melhor organização utilize o template [Projeto Gorote](https://github.com/ronaldalds/gorote):
```
/gorote
├── /api             	# Início do projeto instancia do fiber e configurações
├── /app             	# Pasta onde fica as aplicações do projeto
├── .env.example     	# Exemplo de .env
├── .gitignore        	# Git ignore para go
├── docker-compose.yaml	# docker compose com todos os serviços para subir aplicação em modo dev
├── Dockerfile       	# Dockerfile para iniciar o container da API
├── go.mod           	# Dependências do Go
├── private_key.pem  	# Chave privada (não versionar!) ***Necessário criar***
└── public_key.pem   	# Chave pública			 ***Necessário criar***
```

## 🔐 Serviço de Autenticação Principal

O serviço principal é responsável por:

  - Gerenciar usuários
  - Gerar tokens JWT
  - Validar credenciais
  - Gerenciar tokens de refresh

### Configuração mínima

Crie um arquivo `main.go`:
```go
package main

import (
	"log"
	"time"

	"github.com/ronaldalds/gorote-core-rsa/core"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Gorote Auth Service",
	})

	// 1. Carregar chave privada
	privateKey, err := core.ReadRSAPrivateKeyFromFile("private_key.pem")
	if err != nil {
		log.Fatal("Falha ao ler chave privada:", err)
	}

	// 2. Configuração do banco de dados
	// recomendado usar variáveis de ambiente para maior segurança
	dbConfig := &core.InitGorm{
		Host:     "localhost",			
		User:     "seu_usuario",
		Password: "sua_senha",
		Database: "gorote",
		Port:     5432,
		Schema:   "public",
		TimeZone: "America/Sao_Paulo",
	}

	// 3. Configuração JWT
	jwtConfig := &core.AppJwt{
		JwtExpireAccess:  5 * time.Minute,   // Token de acesso
		JwtExpireRefresh: 24 * time.Hour,    // Token de refresh
	}

	// 4. Super usuário (criado na primeira execução)
	superUser := &core.AppSuper{
		SuperName:  "Admin",
		SuperUser:  "admin",
		SuperPass:  "senha_segura",
		SuperEmail: "admin@empresa.com",
		SuperPhone: "+5511999999999",
	}

	// 5. Configuração completa do app
	appConfig := &core.AppConfig{
		App:         app,
		GormStore:    core.NewGormStore(dbConfig),
		AppName:     "Gorote Auth",
		AppTimeZone: "America/Sao_Paulo",
		PrivateKey:  privateKey,
		Jwt:         jwtConfig,
		Super:       superUser,
	}

	// 6. Inicializar rotas
	router := core.New(appConfig)
	router.RegisterRouter(app.Group("/api/v1"))

	// 7. Iniciar servidor
	log.Fatal(app.Listen(":3000"))
}
```

## 🌐 Microserviço Exemplo

Como consumir a autenticação em outros serviços:
```go

package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/core"
	"github.com/ronaldalds/gorote-core-rsa/example"
)

func main() {
	app := fiber.New()

	// 1. Carregar chave pública
	publicKey, err := core.ReadRSAPublicKeyFromFile("public_key.pem")
	if err != nil {
		log.Fatal("Falha ao ler chave pública:", err)
	}

	// 2. Configuração do banco do microserviço
	serviceDB := &core.InitGorm{
		Host:     "localhost",
		User:     "seu_usuario",
		Password: "sua_senha",
		Database: "service",
		Port:     5432,
	}

	// 3. Configuração do microserviço
	serviceConfig := &example.AppConfig{
		App:         app,
		GormStore:   core.NewGormStore(serviceDB),
		AppName:     "Meu Microserviço",
		PublicKey:   publicKey,
	}

	// 4. Registrar rotas
	service := example.New(serviceConfig)
	service.RegisterRouter(app.Group("/api/v1"))

	log.Fatal(app.Listen(":3001"))
}
```

## 🔍 Endpoints Disponíveis
### Autenticação
| Método | Endpoint             | Descrição                     | Body Request Example             |
|--------|----------------------|-------------------------------|----------------------------------|
| `GET`  |`/api/v1/health`      | Faz um health check           |                                  |
| `POST` |`/api/v1/auth/login`  | Login de usuário              |```{"username":"admin", "password":"senha123"}``` |
| `POST` |`/api/v1/refresh`     | Renova o token de acesso      |```{"refresh_token": "token"}``` |

### Microserviço
| Método | Endpoint             | Descrição                     | Body Request Example             |
|--------|----------------------|-------------------------------|----------------------------------|
| `POST` |`/api/v1/example`     | Sua rota                      |```{"example":"example","example":"example"}``` |


## 🛡️ Como a autenticação funciona
- **Login:**
  - Usuário envia credenciais para `/auth/login`
  - Serviço valida e retorna:
    - `access_token` (validade curta)
    - `refresh_token` (validade longa)

- **Acesso a microserviços:**
  - Incluir header: `Authorization: Bearer <access_token>`
  - Microserviço valida assinatura com chave pública

- **Token expirado:**
  - Client usa `/auth/refresh` com `refresh_token`
  - Recebe novo `access_token`

## 📦 Estrutura do Token JWT
```json
{
  "sub": "123",				// ID do usuário
  "iss": "Gorote",			// Nome App
  "permissions": ["view", "create"],	// Permissões do usuário
  "isSuperUser": false,			// Se é super usuário
  "iat": 1516239022,			// Emitido em
  "exp": 1516242622			// Expira em
}
```

## 🧪 Testando
### 1. Login
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"senha_segura"}'
```

### 2. Acessar microserviço
```bash
curl http://localhost:3001/api/v1/ \
  -H "Authorization: Bearer <SEU_ACCESS_TOKEN>"
```

## 🚨 Segurança

- **Nunca versionar `private_key.pem`**  
  Mantenha este arquivo fora do controle de versão (adicione ao `.gitignore`)

- **Use HTTPS em produção**  
  Sempre habilite SSL/TLS para todas as comunicações

- **Configure tempos de expiração adequados**  
  - Access tokens: 5-15 minutos (ex: `5` em minutos)
  - Refresh tokens: 7-30 dias (ex: `168` em minutos) 

- **Revise permissões do banco de dados**  
  Aplique o princípio do menor privilégio para usuários do DB

- **Monitore tentativas de login**  
  Implemente logs e alertas para múltiplas falhas de autenticação

## ✉️ Contato

**Ronald Almeida** - Desenvolvedor & Mantenedor  
📧 Email: [ronald.ralds@gmail.com](mailto:ronald.ralds@gmail.com)  
💼 LinkedIn: [Ronald Almeida](https://www.linkedin.com/in/ronald-ralds) 

**Repositório do Projeto**:  
📦 GitHub: [github.com/ronaldalds/gorote-core-rsa](https://github.com/ronaldalds/gorote-core-rsa)  

**Relatar Issues**:  
📬 [Issues do Projeto](https://github.com/ronaldalds/gorote-core-rsa/issues)  
