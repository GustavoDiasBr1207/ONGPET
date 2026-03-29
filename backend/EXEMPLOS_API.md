#!/usr/bin/env markdown
# 🎯 GUIA VISUAL COM EXEMPLOS REAIS

## 📡 REQUISIÇÃO 1: Verificar Configuração

### REQUEST
```bash
curl http://localhost:8080/api/v1/test/email-config
```

### RESPONSE ✅ (Sucesso)
```json
{
  "status": "ok",
  "message": "Email configurado corretamente",
  "from": "gustavodl.gdl33@gmail.com"
}
```

### RESPONSE ❌ (Erro)
```json
{
  "status": "error",
  "message": "Mailer não inicializado"
}
```

**Solução**: Verificar logs da aplicação, pode ser erro no SMTP_PASS

---

## 📡 REQUISIÇÃO 2: Enviar Email de Teste

### REQUEST
```bash
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "adoption_request",
    "data": {
      "pet_name": "Buddy",
      "requester_name": "João Silva",
      "ong_name": "ONG Paws"
    }
  }'
```

### RESPONSE ✅ (Sucesso)
```json
{
  "message": "Email de teste enviado com sucesso",
  "to": "gustavodl.gdl33@gmail.com",
  "template": "adoption_request"
}
```

### EMAIL RECEBIDO 📧
```
From: gustavodl.gdl33@gmail.com
To: gustavodl.gdl33@gmail.com
Subject: 🐾 Nova solicitação de adoção: Buddy

──────────────────────────────────────────
        Nova Solicitação de Adoção

Olá ONG Paws,

Você recebeu uma nova solicitação de adoção para o pet Buddy.

Solicitante: João Silva

Acesse o painel da ONG para visualizar os detalhes da solicitação.

Atenciosamente,
Sistema OngPet
──────────────────────────────────────────
```

---

## 📡 REQUISIÇÃO 3: Criar Pedido de Adoção (Envia Email Automático)

### REQUEST
```bash
curl -X POST http://localhost:8080/api/v1/pedidos-adocao \
  -H "Content-Type: application/json" \
  -d '{
    "ong_id": "550e8400-e29b-41d4-a716-446655440000",
    "pet_id": "550e8400-e29b-41d4-a716-446655440001",
    "respostas": [
      {
        "campo_formulario_id": "550e8400-e29b-41d4-a716-446655440002",
        "valor": "Maria Silva"
      },
      {
        "campo_formulario_id": "550e8400-e29b-41d4-a716-446655440003",
        "valor": "maria@email.com"
      }
    ]
  }'
```

### RESPONSE ✅ (Imediata - HTTP 201)
```json
{
  "message": "Pedido de adoção criado com sucesso",
  "pedido": {
    "id": "550e8400-e29b-41d4-a716-446655440004",
    "ong_id": "550e8400-e29b-41d4-a716-446655440000",
    "pet_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "pendente",
    "created_at": "2026-03-29T10:30:00Z",
    "respostas": [...]
  }
}
```

### LOG DA APLICAÇÃO 👀 (Segundos depois)
```
✅ Email enviado para ONG: suaong@example.com
```

### EMAIL RECEBIDO NA ONG 📧 (1-2 minutos depois)
```
From: gustavodl.gdl33@gmail.com
To: suaong@example.com
Subject: 🐾 Nova solicitação de adoção: [Nome do Pet]

──────────────────────────────────────────
        Nova Solicitação de Adoção

Olá [Nome ONG],

Você recebeu uma nova solicitação de adoção para o pet [Nome Pet].

Solicitante: Maria Silva

Acesse o painel da ONG para visualizar os detalhes da solicitação.

Atenciosamente,
Sistema OngPet
──────────────────────────────────────────
```

---

## 📈 FLUXO VISUAL COMPLETO

```
┌─────────────────────────────┐
│    CLIENTE FAZ REQUISIÇÃO   │
│  POST /api/v1/pedidos-adocao │
└──────────────┬──────────────┘
               │
               ▼
        ┌──────────────┐
        │   BACKEND    │
        └──────┬───────┘
               │
        ┌─────▼─────────────────────────────┐
        │ 1. Valida dados de entrada         │
        │ 2. Cria registro no DB             │
        │ 3. Retorna HTTP 201 ◄── CLIENTE   │
        │    VIRA PARA USUÁRIO               │
        └─────┬─────────────────────────────┘
              │
      [GOROUTINE - Background]
      [NÃO BLOQUEIA MAIS NADA]
              │
        ┌─────▼─────────────────────────────┐
        │ 4. Busca dados ONG do DB           │
        │ 5. Monta template email em HTML    │
        │ 6. Envia via SMTP Gmail            │
        └─────┬─────────────────────────────┘
              │
              ▼
       ┌─────────────────┐
       │  SMTP Gmail     │
       │ smtp.gmail.com  │
       │  Port 587 TLS   │
       └────────┬────────┘
                │
                ▼
        ┌──────────────────┐
        │  EMAIL DA ONG    │
        │  suaong@ong.com  │ ✅
        └──────────────────┘
```

---

## 🔍 COMO DEBUGAR

### Se Email NÃO Chegar

**Passo 1: Ver Logs da Aplicação**
```bash
# Procure por essas mensagens no console:
# ✅ Email enviado para ONG: seu_email@gmail.com
# ⚠️ Erro ao enviar email: [motivo]
```

**Passo 2: Verificar Configuração**
```bash
curl http://localhost:8080/api/v1/test/email-config

# Se retornar erro, mailer não está inicializado
```

**Passo 3: Verificar .env**
```bash
grep "^SMTP" .env

# Deve ter exatamente:
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USER=gustavodl.gdl33@gmail.com
# SMTP_PASS=xxxx xxxx xxxx xxxx (16 caracteres)
# SMTP_FROM=gustavodl.gdl33@gmail.com
```

**Passo 4: Testar Conexão**
```bash
telnet smtp.gmail.com 587
# Se conectar, digitar: quit

# Se não conectar:
# - Problema de firewall
# - Problema de DNS
# - Problema de internet
```

**Passo 5: Verificar Pasta de SPAM**
```
Gmail pode marcar na primeira vez como spam
- Procure em: Spam/Lixo
- Clique "Marcar como não spam"
- Adicione remetente aos contatos
```

---

## 📊 TEMPLATES DISPONÍVEIS

### 1. Pet Registered
```bash
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "pet_registered",
    "data": {
      "pet_name": "Fluffy",
      "owner_name": "Maria"
    }
  }'
```

**Email Preview:**
```
Subject: 🐾 Pet registered: Fluffy

Hi Maria!

Your pet Fluffy was successfully registered.
```

---

### 2. Adoption Request
```bash
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "suaong@example.com",
    "template": "adoption_request",
    "data": {
      "pet_name": "Buddy",
      "requester_name": "João Silva",
      "ong_name": "ONG Paws"
    }
  }'
```

**Email Preview:**
```
Subject: 🐾 Nova solicitação de adoção: Buddy

Novo Solicitação de Adoção

Olá ONG Paws,

Você recebeu uma nova solicitação de adoção para o pet Buddy.

Solicitante: João Silva

Acesse o painel da ONG para visualizar os detalhes da solicitação.
```

---

### 3. Adoption Confirmed
```bash
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "joao@email.com",
    "template": "adoption_confirmed",
    "data": {
      "pet_name": "Buddy",
      "requester_name": "João Silva"
    }
  }'
```

**Email Preview:**
```
Subject: 🎉 Adoção aprovada: Buddy

Parabéns! Sua adoção foi aprovada!

Olá João Silva,

Sua solicitação de adoção para o pet Buddy foi aprovada!

A ONG entrará em contato com você em breve para os próximos passos.
```

---

### 4. Contact
```bash
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "admin@ong.com",
    "template": "contact",
    "data": {
      "name": "Pedro",
      "email": "pedro@email.com",
      "message": "Gostaria de adotar um pet..."
    }
  }'
```

**Email Preview:**
```
Subject: 📧 Novo contato: Pedro

Novo Contato

Nome: Pedro
Email: pedro@email.com

Mensagem:
Gostaria de adotar um pet...
```

---

## ✅ VERIFICATIONS CHECKLIST

Após configurar tudo, marque cada item:

- [ ] `go get gopkg.in/gomail.v2` rodou sem erros
- [ ] SMTP_PASS foi adicionado ao .env
- [ ] `go run main.go` mostra "✅ Mailer inicializado"
- [ ] GET /api/v1/test/email-config retorna "status": "ok"
- [ ] POST /api/v1/test/email envia email com sucesso
- [ ] Email aparece em gustavodl.gdl33@gmail.com
- [ ] Criar pedido de adoção dispara email automático
- [ ] Email da ONG contém informações corretas

---

## 🎉 Status Final

```
╔═══════════════════════════════════════════════════════════╗
║                    ✅ TUDO PRONTO!                        ║
║                                                           ║
║  Você pode começar a:                                     ║
║  ✅ Enviar emails de teste                               ║
║  ✅ Criar pedidos com email automático                   ║
║  ✅ Customizar templates (opcional)                      ║
║  ✅ Monitorar em produção                                ║
║                                                           ║
║  Email configurado: gustavodl.gdl33@gmail.com             ║
║  Responsável: Gustavo Dias                               ║
║  Data: 29 de Março de 2026                               ║
╚═══════════════════════════════════════════════════════════╝
```
