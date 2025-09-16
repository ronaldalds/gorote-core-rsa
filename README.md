# Gorote Core RSA - Authentication Library

![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)
![Fiber](https://img.shields.io/badge/Fiber-2.x-00ADD8)
![JWT](https://img.shields.io/badge/JWT-RSA-000000?logo=JSON%20web%20tokens)
[![Go Report Card](https://goreportcard.com/badge/github.com/ronaldalds/gorote-core-rsa)](https://goreportcard.com/report/github.com/ronaldalds/gorote-core-rsa)

Uma biblioteca completa para autenticação e autorização usando JWT com criptografia RSA, desenvolvida em Go com Fiber.

## 📋 Pré-requisitos

- Go 1.24 ou superior
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
Configuração necessária caso não existam os database.

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
├── /env             	# Pasta onde env do projeto
├── .env.example     	# Exemplo de .env
├── .gitignore        	# Git ignore para go
├── docker-compose.yaml	# docker compose com todos os serviços para subir aplicação em modo dev
├── Dockerfile       	# Dockerfile para iniciar o container da API
├── go.mod           	# Dependências do Go
├── go.sum           	# Dependências do Go
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
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/core"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"gorm.io/driver/postgres"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Configurar variáveis da aplicação
	appName := "core"
	appVersion := "v0.1.0"
	appTimezone := "America/Fortaleza"
	appPort := 3000
	domain := "localhost"

	// 2. Configuração do banco de dados
	// recomendado usar variáveis de ambiente para maior segurança
	dbConfig := &gorote.InitPostgres{
		Host:     "localhost",
		User:     "seu_usuario",
		Password: "sua_senha",
		Database: "gorote",
		Port:     5432,
		Schema:   "public",
		TimeZone: appTimezone,
	}
	// iniciar GORM
	sql, err := gorote.NewGorm(postgres.Open(dbConfig.DSN()))
	if err != nil {
		log.Fatal("err on sql")
	}

	// 3. Configuração JWT
	jwtExpireAccess := 5 * time.Minute // Token de acesso
	jwtExpireRefresh := 24 * time.Hour // Token de refresh

	// 4. Super usuário (criado na primeira execução)
	superEmail := "admin@admin.com"
	superPass := "admin"

	// 5. Carregar chave privada
	privateKey := gorote.MustReadPrivateKeyFromFile("private_key.pem")

	// 6. Iniciar fiber server
	app := fiber.New(fiber.Config{
		AppName: appName,
	})

	// 7. Configuração da url do colletor para uso do opentelemetry
	// caso nao use telemetria não será necessario
	collectorOpentelemetry := "localhost:4317"
	// configurar resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(appName),
			semconv.ServiceVersionKey.String(appVersion),
		),
	)
	if err != nil {
		log.Fatal("error on resource")
	}
	// iniciar provider
	if err := gorote.TelemetryFiber(ctx, app, res, collectorOpentelemetry); err != nil {
		log.Fatal("error on telemetry")
	}

	// 8. Configuração completa do app
	coreRouter, err := core.New(&core.Config{
		DB:               sql,
		AppName:          appName,
		PrivateKey:       privateKey,
		JwtExpireAccess:  jwtExpireAccess,
		JwtExpireRefresh: jwtExpireRefresh,
		SuperEmail:       superEmail,
		SuperPass:        superPass,
		Domain:           domain,
	})
	if err != nil {
		log.Fatal("err on config core")
	}

	// 9. Inicializar rotas
	coreRouter.RegisterRouter(app.Group("/api/v1"))

	// 10. Iniciar servidor
	go func() {
		err := app.Listen(fmt.Sprintf(":%d", appPort))
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	<-ctx.Done()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(contx); err != nil {
		log.Fatal("server forced to shutdown with error")
	}
}
```

## 🌐 Microserviço Exemplo

Como consumir a autenticação em outros serviços:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/example"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"gorm.io/driver/postgres"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Configurar variáveis da aplicação
	appName := "example"
	appVersion := "v0.1.0"
	appTimezone := "America/Fortaleza"
	appPort := 3000

	// 2. Configuração do banco de dados
	// recomendado usar variáveis de ambiente para maior segurança
	dbConfig := &gorote.InitPostgres{
		Host:     "localhost",
		User:     "seu_usuario",
		Password: "sua_senha",
		Database: "gorote",
		Port:     5432,
		Schema:   "public",
		TimeZone: appTimezone,
	}
	// iniciar GORM
	sql, err := gorote.NewGorm(postgres.Open(dbConfig.DSN()))
	if err != nil {
		log.Fatal("err on sql")
	}

	// 3. Carregar chave privada
	publicKey := gorote.MustReadPublicKeyFromFile("public_key.pem")

	// 4. Iniciar fiber server
	app := fiber.New(fiber.Config{
		AppName: appName,
	})

	// 5. Configuração da url do colletor para uso do opentelemetry
	// caso nao use telemetria não será necessario
	collectorOpentelemetry := "localhost:4317"
	// configurar resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(appName),
			semconv.ServiceVersionKey.String(appVersion),
		),
	)
	if err != nil {
		log.Fatal("error on resource")
	}
	// iniciar provider
	if err := gorote.TelemetryFiber(ctx, app, res, collectorOpentelemetry); err != nil {
		log.Fatal("error on telemetry")
	}

	// 6. Configuração completa do app
	microRouter, err := example.New(&example.Config{
		DB:        sql,
		PublicKey: publicKey,
	})
	if err != nil {
		log.Fatal("err on config micro")
	}

	// 7. Inicializar rotas
	microRouter.RegisterRouter(app.Group("/api/v1"))

	// 8. Iniciar servidor
	go func() {
		err := app.Listen(fmt.Sprintf(":%d", appPort))
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	<-ctx.Done()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(contx); err != nil {
		log.Printf("server forced to shutdown with error")
	}
}
```

## 🔍 Endpoints Disponíveis
### Autenticação
| Método | Endpoint             | Descrição                     | Body Request Example             |
|--------|----------------------|-------------------------------|----------------------------------|
| `GET`  |`health`      | Faz um health check           |                                  |
| `POST` |`auth/login`  | Login de usuário              |```{"email":"admin@admin.com", "password":"admin"}``` |
| `POST` |`refresh`     | Renova o token de acesso      |```{"refresh_token": "token"}``` |

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
  "isSuperUser": true,
  "permissions": ["string_permission","string_permission"],
  "tenants": ["uuid","uuid"],
  "type": "access_token",
  "iss": "app_name",
  "sub": "admin@admin.com",
  "exp": 1757970475,
  "iat": 1757970175,
  "jti": "d92c1253-af18-4d9a-7164-f8d488671779"
}
```

## 🧪 Testando
### 1. Login
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@admin.com","password":"admin"}'
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
  - Access tokens: 5-15 minutos (ex: `300` em segundos)
  - Refresh tokens: 7-30 dias (ex: `604800` em segundos) 

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
