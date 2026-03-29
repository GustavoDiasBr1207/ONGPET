# 🏗️ Arquitetura do Sistema de Email

## Fluxo Completo

```
┌─────────────────────────────────────────────────────────────────┐
│                     APLICAÇÃO ONGPET                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MAIN.GO (Startup)                           │
│  ├─ utils.InitMailer()  ◄── Inicializa conexão SMTP            │
│  └─ Carrega .env com SMTP_* config                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              MAILER_SERVICE.GO (Singleton)                       │
│  ├─ GetMailer() - Retorna instância global                      │
│  ├─ SendEmailPetRegistered()                                    │
│  ├─ SendEmailAdoptionRequest()                                  │
│  ├─ SendEmailAdoptionConfirmed()                                │
│  └─ SendEmailContact()                                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              EMAIL.GO (Mailer Struct)                            │
│  ├─ type Mailer struct                                          │
│  ├─ New() - Conecta ao SMTP                                     │
│  └─ Send(to, subject, body) - Envia email                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              TEMPLATES.GO (HTML Templates)                       │
│  ├─ NewPetRegisteredEmail()                                     │
│  ├─ NewAdoptionRequestEmail()                                   │
│  ├─ NewAdoptionConfirmedEmail()                                 │
│  └─ NewContactEmail()                                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  GOPKG.IN/GOMAIL.V2                              │
│           (Biblioteca de envio SMTP)                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   SMTP.GMAIL.COM:587                             │
│           (Servidor Gmail com TLS)                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  CAIXA DE ENTRADA DO USUÁRIO                    │
└─────────────────────────────────────────────────────────────────┘
```

## Fluxo de Criação de Pedido de Adoção com Email

```
┌──────────────┐
│  API Client  │
└──────────────┘
       │
       │ POST /api/v1/pedidos-adocao
       ▼
┌────────────────────────────────────────────────┐
│  CreatePedidoAdocao (pedidoadocao.go)           │
│                                                 │
│  1. Valida dados de entrada                     │
│  2. Cria registro pedido no DB                  │
│  3. Cria respostas do formulário                │
└────────────────────────────────────────────────┘
       │
       ├─► Retorna resposta HTTP 201 ◄─── Cliente recebe imediatamente
       │
       │ (GOROUTINE - Não bloqueia)
       ▼
┌────────────────────────────────────────────────┐
│  SendEmailAdoptionRequest()                    │
│  (background - não interfere na resposta)      │
│                                                 │
│  1. Busca dados da ONG                          │
│  2. Busca nome do solicitante nas respostas    │
│  3. Gera template do email                      │
│  4. Envia para email da ONG                    │
└────────────────────────────────────────────────┘
       │
       ▼
   ✅ Email recebido na caixa de entrada da ONG
```

## Componentes Principais

### 1️⃣ EMAIL.GO
```go
type Mailer struct {
    dialer *gomail.Dialer
    from   string
}

func (m *Mailer) Send(to []string, subject, body string) error
```
- Mantém conexão com servidor SMTP
- Envia emails em formato HTML
- Tratamento de erros centralizado

### 2️⃣ MAILER_SERVICE.GO
```go
func InitMailer() error              // Setup único
func GetMailer() *Mailer             // Instância global
func SendEmailPetRegistered(...)     // Pet registrado
func SendEmailAdoptionRequest(...)   // Nova solicitação
func SendEmailAdoptionConfirmed(...) // Adoção aprovada
func SendEmailContact(...)           // Contato geral
```
- Padrão Singleton para mailer global
- Funções de conveniência para cada tipo de email
- Tratamento de falhas sem prejudicar operação principal

### 3️⃣ TEMPLATES.GO
```go
func NewPetRegisteredEmail(petName, ownerName string) (subject, body string)
func NewAdoptionRequestEmail(petName, requesterName, ongName string) (subject, body string)
func NewAdoptionConfirmedEmail(petName, requesterName string) (subject, body string)
func NewContactEmail(name, email, message string) (subject, body string)
```
- Templates em HTML formatados
- Fácil manutenção e customização
- Suporte a variáveis dinâmicas

### 4️⃣ PEDIDOADOCAO.GO (Controllers)
```go
func CreatePedidoAdocao(c *gin.Context) error {
    // ... criar pedido ...
    
    go func() {
        // Enviar email em background
        utils.SendEmailAdoptionRequest(...)
    }()
}
```
- Integração automática no fluxo de criação
- Uso de goroutine para não bloquear resposta

### 5️⃣ EMAIL_TEST.GO (Controllers)
```go
func SendTestEmail(c *gin.Context) error     // POST /api/v1/test/email
func CheckEmailConfig(c *gin.Context) error  // GET /api/v1/test/email-config
```
- Endpoints para validação da configuração
- Testes manuais via API
- Útil para troubleshooting

## Configuração (.env)

```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=gustavodl.gdl33@gmail.com
SMTP_PASS=sua_senha_de_app (16 caracteres)
SMTP_FROM=gustavodl.gdl33@gmail.com
```

## Tratamento de Erros

```
┌─────────────────────────────────────────────────┐
│  Erro ao enviar email                            │
└─────────────────────────────────────────────────┘
       │
       ├─► Sistema log do erro ⚠️
       │
       ├─► NÃO interrompe operação principal ✅
       │
       └─► Retorna `nil` (falha silenciosa)
           (não bloqueia requisição HTTP)
```

## Performance

- ✅ Emails enviados em goroutines (não-bloqueante)
- ✅ Singleton Mailer (reutilização de conexão)
- ✅ HTML compilado em memória (sem arquivo)
- ✅ TLS 587 (encriptado e rápido)
- ⏱️ Tempo típico de envio: 100-500ms

## Segurança

- 🔒 Senha armazenada em .env (não no código)
- 🔐 TLS/transport encryption (SMTP 587)
- 🚫 Remoção de variáveis sensíveis de logs
- ✅ Validação de email em entrada (controllers/v1/email_test.go)
