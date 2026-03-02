# Shorty 🔗⚡️

Shorty é um encurtador de URLs desenvolvido em Go, focado em arquitetura limpa, alta performance e processamento assíncrono de eventos.

O projeto foi construído com mentalidade de sistemas distribuídos reais, separando responsabilidades entre API HTTP e Workers assíncronos utilizando Kafka.

---

## Features

- API REST em Go usando Fiber
- Autenticação JWT
- Criação de URLs encurtadas
- Redirect público ultra rápido
- Cache com Redis (opcional)
- Eventos assíncronos via Kafka
- Worker para processamento de analytics
- PostgreSQL como source of truth
- Arquitetura modular (Clean Architecture inspired)

---

## Arquitetura

O sistema roda em dois modos:

### API Mode (`APP_MODE=api`)

Responsável por:

- Servir API HTTP
- Autenticação
- Criar short URLs
- Realizar redirects
- Publicar eventos de clique no Kafka
- Utilizar cache Redis quando disponível

Fluxo do redirect:

Client -> API -> Redis (cache hit)
                -> Postgres (fallback)
                -> Kafka event (async)
                -> Redirect

---

### Worker Mode (`APP_MODE=worker`)

Responsável por:

- Consumir eventos Kafka (`url.clicks`)
- Processar analytics
- Incrementar contador de visitas no banco

Separar worker da API evita bloquear requests HTTP.

---

## Stack Tecnológica

- Go
- Fiber v2
- PostgreSQL (pgx)
- Redis
- Kafka (Sarama)
- JWT Authentication
- Docker Compose

---

## Requisitos

- Go instalado
- Docker
- Docker Compose
- migrate CLI (golang-migrate)

Instalar migrate:

Mac:
brew install golang-migrate

Linux:
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz

---

## Rodando Localmente

### 1. Subir infraestrutura

docker compose up -d

Serviços disponíveis:

Postgres -> localhost:5432  
Redis -> localhost:6379  
Kafka -> localhost:9092  
Kafka UI -> http://localhost:8090  

---

### 2. Executar migrations

make migrate-up

Rollback:

make migrate-down

Criar nova migration:

make migrate-create name=create_short_urls

---

### 3. Rodar API

make run

ou

APP_MODE=api go run ./cmd/api-server

Servidor inicia em:

http://localhost:8080

---

### 4. Rodar Worker

em outro terminal:

APP_MODE=worker go run ./cmd/api-server

---

## Variáveis de Ambiente

Criar arquivo `.env`

APP_PORT=8080
APP_MODE=api
BASE_URL=http://localhost:8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=shorty
DB_PASSWORD=shorty
DB_NAME=shorty
DB_SSLMODE=disable

JWT_SECRET=super-secret-key

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

KAFKA_ENABLED=true
KAFKA_BROKERS=localhost:9092

WORKER_GROUP_ID=shorty-workers
WORKER_TOPICS=url.clicks

---

## Endpoints

### Health Check

GET /health

Retorna status da aplicação e conexões.

---

### Auth

#### Register

POST /api/v1/auth/register

Body:

{
  "email": "user@email.com",
  "password": "12345678"
}

Regras:

- email validado
- senha mínima 8 caracteres
- hash bcrypt

---

#### Login

POST /api/v1/auth/login

Retorna:

- accessToken
- refresh token em cookie HttpOnly

---

### Criar Short URL

POST /api/v1/shorturls

Header:

Authorization: Bearer TOKEN

Body:

{
  "url": "https://example.com",
  "expires_at": "2026-12-31T23:59:59Z"
}

Resposta:

{
  "short_code": "AbCdE12",
  "short_url": "http://localhost:8080/AbCdE12"
}

---

### Redirect Público

GET /:code

Fluxo:

1. tenta Redis
2. fallback Postgres
3. publica evento Kafka
4. retorna redirect 302

---

### Usuário Atual

GET /api/v1/me

Retorna dados do JWT autenticado.

---

## Kafka Events

### Topics

url.clicks -> analytics de acesso  
url.created -> futuro  
url.expired -> futuro  

Evento contém:

- EventID
- OccurredAt
- ShortCode
- IP
- UserAgent
- Referer

Key Kafka = shortCode

---

## Estrutura do Projeto

cmd/
  api-server/

internal/
  cache/
  config/
  db/
  domain/
  events/
  handler/
  http/
  queue/
  repository/postgres/
  security/jwt/
  security/refresh/
  service/
  worker/

migrations/
docker-compose.yml
Makefile
go.mod

---

## Decisões de Design

- Redis é opcional (fail-open)
- Kafka publish é best-effort (não bloqueia redirect)
- Worker separado melhora escalabilidade
- Postgres permanece como fonte única de verdade
- Cache apenas acelera leitura

---

## Roadmap

- Refresh token endpoint
- Listagem paginada de URLs
- Rate limiting
- Observabilidade (OpenTelemetry)
- Métricas Prometheus
- Docker multi-stage build
- Deploy Kubernetes

---

## Licença

Copyright (c) 2026 Matheus Parro

All rights reserved.

This repository is provided for viewing and educational purposes only.

No permission is granted to copy, modify, distribute, sublicense, or use
this software or its source code in any commercial or non-commercial project
without explicit written permission from the author.