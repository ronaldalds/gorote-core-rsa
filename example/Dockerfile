# Etapa 1: imagem base com Go para compilar o código
FROM golang:1.24-bookworm as builder

# Cria diretório do projeto e define como diretório de trabalho
WORKDIR /app

# Copia os arquivos do projeto Go
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Compila o binário
RUN go build -o main .

# Etapa 2: imagem final mais leve
FROM debian:bookworm-slim

# Instala apenas o Chromium para rodar o navegador headless
RUN apt-get update && \
    apt-get install -y chromium ca-certificates && \
    apt-get clean

# Copia o binário gerado na etapa anterior
COPY --from=builder /app/main /

# Copia o Chromium também, se necessário (opcional)

# Define a variável de ambiente para o Rod encontrar o navegador
ENV ROD_BROWSER_PATH=/usr/bin/chromium

# Executa o binário
CMD "/main"
