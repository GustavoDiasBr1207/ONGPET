# OngPet — Backend

API REST para plataforma de adoção de animais. Permite que ONGs cadastrem pets, gerenciem pedidos de adoção, enviem notificações por e-mail e WhatsApp, e façam acompanhamento pós-adoção.

## Stack

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.24 |
| Framework HTTP | Gin |
| ORM | GORM |
| Banco de dados | PostgreSQL (Supabase) |
| Autenticação | Supabase Auth (JWT ES256 via JWKS) |
| Storage | Supabase Storage |
| E-mail | SendGrid API |
| WhatsApp | Evolution API |
| Documentação | Swagger (swaggo) |
| Deploy | Docker / Render |

---

## Funcionalidades

- **ONGs** — CRUD completo com upload de logo
- **Pets** — CRUD com múltiplas imagens, filtros por espécie, porte, região e status
- **Banners** — gerenciados por admins, com imagem
- **Formulários** — modelos de formulário com campos dinâmicos e upload de imagem por resposta
- **Pedidos de adoção** — envio público com rate-limit, workflow de status (pendente → aprovado/recusado)
- **Acompanhamento pós-adoção** — frequência configurável (único, mensal, trimestral, semestral) com logs e lembretes automáticos
- **Notificações** — e-mail via SendGrid e WhatsApp via Evolution API (ambos opcionais via env)
- **Swagger** — documentação interativa habilitável via env

---

## Variáveis de Ambiente

Copie `.envmodel.txt` para `.env` e preencha:

```env
# ─── Banco de Dados ──────────────────────────────────────────────────────────
DATABASE_URL=postgresql://...

# ─── Supabase ────────────────────────────────────────────────────────────────
SUPABASE_URL=https://<project>.supabase.co
SUPABASE_KEY=<anon/publishable key>
SUPABASE_JWT_SECRET=<jwt secret>

# ─── Supabase Storage Buckets ────────────────────────────────────────────────
SUPABASE_BUCKET_PETS=pets-images
SUPABASE_BUCKET_ONGS=OngImages
SUPABASE_BUCKET_FORMULARIOS=FormImages
SUPABASE_BUCKET_BANNERS=BannerImages

# ─── Email (SendGrid) ────────────────────────────────────────────────────────
SMTP_PASS=SG.<sua-api-key>
SMTP_FROM=seuemail@dominio.com
TEST_EMAIL_TO=                   # opcional — destino dos testes de email

# ─── WhatsApp (Evolution API) ────────────────────────────────────────────────
WHATSAPP_ENABLED=false           # true para habilitar
WHATSAPP_API_URL=http://evolution:8080/message/sendText/AdocaoPet
WHATSAPP_API_KEY=<api-key>
WHATSAPP_SENDER_NUMBER=55279...
TEST_WHATSAPP_NUMBER=55279...

# ─── Administração ───────────────────────────────────────────────────────────
ADMIN_USER_IDS=<uuid1>,<uuid2>   # UUIDs separados por vírgula (Supabase Auth)

# ─── Swagger ─────────────────────────────────────────────────────────────────
SWAGGER_ENABLED=false            # true para habilitar /swagger/index.html

# ─── Debug ───────────────────────────────────────────────────────────────────
DEBUG=false
GIN_MODE=release                 # release | debug
```

> `ADMIN_USER_IDS` — UUIDs dos usuários admin encontrados em **Supabase → Authentication → Users**. Se não configurado, qualquer usuário autenticado acessa rotas de admin (com aviso no log).

---

## Rodando Localmente

**Pré-requisitos:** Go 1.24+, PostgreSQL acessível (ou usar Supabase).

```bash
# Clone e instale dependências
git clone <repo>
cd backend

# Configure o ambiente
cp .envmodel.txt .env
# edite .env com suas credenciais

# Gere a documentação Swagger (opcional)
swag init -g main.go -o docs

# Execute
go run .
```

A API sobe em `http://localhost:8080`.  
Swagger disponível em `http://localhost:8080/swagger/index.html` quando `SWAGGER_ENABLED=true`.

---

## Docker Compose

O `docker-compose.yml` sobe a API junto com a **Evolution API** (WhatsApp) e suas dependências (PostgreSQL + Redis):

```bash
docker compose up -d
```

| Serviço | Porta local | Descrição |
|---|---|---|
| `app` | 8080 | API OngPet |
| `evolution` | 8081 | Evolution API (WhatsApp) |
| `postgres-evolution` | 5433 | PostgreSQL da Evolution |
| `redis` | 6379 | Redis para sessões do Evolution |

> O banco de dados principal (Supabase) é externo e configurado via `DATABASE_URL`.

---

## Endpoints

### Auth
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/api/v1/auth/login` | — | Login via Supabase |

### ONGs
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/api/v1/ongs` | — | Lista ONGs |
| GET | `/api/v1/ongs/:id` | — | Busca ONG por ID |
| POST | `/api/v1/ongs` | ✓ | Cria ONG |
| PUT | `/api/v1/ongs/:id` | ✓ | Atualiza ONG |
| DELETE | `/api/v1/ongs/:id` | ✓ | Remove ONG |
| POST | `/api/v1/ongs/:id/logo` | ✓ | Upload de logo |
| DELETE | `/api/v1/ongs/:id/logo` | ✓ | Remove logo |

### Pets
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/api/v1/pets` | — | Lista pets (paginado, com filtros) |
| GET | `/api/v1/pets/:id` | — | Busca pet por ID |
| POST | `/api/v1/pets` | ✓ | Cria pet |
| PUT | `/api/v1/pets/:id` | ✓ | Atualiza pet |
| DELETE | `/api/v1/pets/:id` | ✓ | Remove pet |
| POST | `/api/v1/pets/:id/imagens` | ✓ | Upload de imagens |
| DELETE | `/api/v1/pets/:id/imagens/:imageId` | ✓ | Remove imagem |

### Banners *(admin)*
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/api/v1/banners` | — | Lista banners |
| GET | `/api/v1/banners/:id` | — | Busca banner |
| POST | `/api/v1/banners` | ✓ admin | Cria banner |
| PUT | `/api/v1/banners/:id` | ✓ admin | Atualiza banner |
| DELETE | `/api/v1/banners/:id` | ✓ admin | Remove banner |
| POST | `/api/v1/banners/:id/imagem` | ✓ admin | Upload de imagem |
| DELETE | `/api/v1/banners/:id/imagem` | ✓ admin | Remove imagem |

### Formulários
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET/POST | `/api/v1/formularios` | —/✓ | Lista / Cria |
| GET/PUT/DELETE | `/api/v1/formularios/:id` | —/✓ | Busca / Atualiza / Remove |
| POST/PUT/DELETE | `/api/v1/formularios/:id/campos/:campoId` | ✓ | Gerencia campos |

### Pedidos de Adoção
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/api/v1/pedidos-adocao` | ✓ | Lista pedidos |
| GET | `/api/v1/pedidos-adocao/:id` | ✓ | Busca pedido |
| POST | `/api/v1/pedidos-adocao` | — | Envia pedido *(rate-limited)* |
| PUT | `/api/v1/pedidos-adocao/:id/status` | ✓ | Atualiza status |
| DELETE | `/api/v1/pedidos-adocao/:id` | ✓ | Remove pedido |

### Acompanhamento Pós-Adoção
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/api/v1/acompanhamentos` | ✓ | Cria acompanhamento |
| GET | `/api/v1/acompanhamentos` | ✓ | Lista acompanhamentos |
| POST | `/api/v1/acompanhamentos/:id/logs` | ✓ | Registra log de contato |
| GET | `/api/v1/acompanhamentos/:id/logs` | ✓ | Lista logs |

### Utilitários *(admin)*
| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/api/v1/test/email-config` | ✓ admin | Verifica configuração de e-mail |
| POST | `/api/v1/test/email` | ✓ admin | Envia e-mail de teste |

---

## Notificações

### E-mail (SendGrid)

Configurado via `SMTP_PASS` e `SMTP_FROM`. Usado para:

- Nova solicitação de adoção recebida pela ONG
- Adoção aprovada / recusada para o solicitante
- Lembretes de acompanhamento pós-adoção

Para testar via terminal:

```bash
go test ./utils/ -run TestSendRealEmail -v
```

### WhatsApp (Evolution API)

Habilitado com `WHATSAPP_ENABLED=true`. Usa a Evolution API como intermediário para enviar mensagens via número conectado. Para rodar localmente com o Docker Compose, o número deve ser vinculado pela interface do Evolution (`http://localhost:8081`).

### Lembretes Automáticos

O scheduler (`utils/scheduler.go`) roda diariamente às **20:00 BRT** e envia lembretes por e-mail e/ou WhatsApp para acompanhamentos cujo `proxima_data` chegou.

---

## Swagger

Para (re)gerar a documentação:

```bash
swag init -g main.go -o docs
```

Para habilitar a interface interativa:

```env
SWAGGER_ENABLED=true
```

Acesse em: `http://localhost:8080/swagger/index.html`

---

## Deploy (Render)

O projeto está pronto para deploy no Render via Docker. Configure as variáveis de ambiente no painel do Render conforme `.envmodel.txt`.

**Build command:** `docker build .`  
**Start command:** `./main`

---

## Testes

```bash
# Todos os testes
go test ./...

# Apenas testes de e-mail
go test ./utils/ -run "TestMailerInit|TestSendRealEmail" -v

# Testes de um controller específico
go test ./controllers/v1/ -run TestPet -v
```
