#!/usr/bin/env markdown
# 📧 SENDGRID - GUIA DE BOAS PRÁTICAS

## 🎯 Sua Configuração Atual

```
📌 Usando: SMTP SendGrid
📌 Email From: amigofieladocao@gmail.com
📌 Status: ✅ Funcionando
```

---

## 🔄 Duas Opções Disponíveis

### Opção 1: SMTP (Atual - Simples)
```
Vantagens:
✅ Funciona com Gomail
✅ Sem bibliotecas extras
✅ Compatível com qualquer provider
✅ Fácil de debugar

Desvantagens:
⚠️ Mais lento para envios em escala
⚠️ Menos control sobre bounces/failures
```

### Opção 2: SendGrid API (Mais Potente)
```
Vantagens:
✅ Mais rápido (HTTP para WebHook)
✅ Rastreamento de opens/clicks
✅ Webhooks de eventos
✅ Melhor para escala alta
✅ Suporte para attachments avançado

Desvantagens:
⚠️ Requer SDK SendGrid
⚠️ Mais complexo de implementar
⚠️ Documentação adicional
```

---

## ✅ Recomendação

**Para seu projeto atual:** Continue com **SMTP (Opção 1)**
- Simples de configurar ✅
- Funciona perfeitamente ✅
- Sem overhead ✅
- Fácil manutenção ✅

**Se precisar de scaling futuro:** Migre para **API SendGrid**
- Melhor performance
- Melhor rastreamento
- Webhooks para eventos

---

## 🚀 Começar a Usar

### 1. Instalar Dependência
```bash
go get -u gopkg.in/gomail.v2
```

### 2. Iniciar Servidor
```bash
go run main.go
```

✅ Log esperado:
```
✅ Mailer inicializado com sucesso
🚀 API rodando em http://localhost:8080
```

### 3. Testar
```bash
# Verificar config
curl http://localhost:8080/api/v1/test/email-config

# Enviar email de teste
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "seu_email@gmail.com",
    "template": "adoption_request",
    "data": {
      "pet_name": "Buddy",
      "requester_name": "João",
      "ong_name": "ONG Amigo Fiel"
    }
  }'
```

---

## 📊 Fluxo de Email com SendGrid

```
sua-api
    ↓
CreatePedidoAdocao (background goroutine)
    ↓
GetMailer() (Singleton)
    ↓
Gomail.Send()
    ↓
SMTP: smtp.sendgrid.net:587
    ↓
SendGrid Infrastructure
    ↓
Email enviado ✅
```

---

## 🔐 Segurança

- ✅ API key em `.env` (não em código)
- ✅ Never commit `.env` com credentials
- ✅ Use `.gitignore` para `.env`
- ✅ Trusted sender: amigofieladocao@gmail.com

### .gitignore (Obrigatório)
```
.env
.env.local
.env.production
```

---

## 📈 Monitoramento

### No Dashboard SendGrid
1. Acesse https://app.sendgrid.com
2. Vá para "Mail Activity"
3. Veja todos os emails enviados
4. Rastreie bounces/complaints

### Via API (Extras)
```bash
# Verificar estatísticas (requer SDK SendGrid)
curl -X GET "https://api.sendgrid.com/v3/stats" \
  -H "Authorization: Bearer SG.Ob3pq..." \
  -H "Content-Type: application/json"
```

---

## 🚨 Troubleshooting SendGrid

### Email não chega?
1. ✅ Verificar no Dashboard SendGrid → Mail Activity
2. ✅ Procurar por "bounces" ou "complaints"
3. ✅ Validar SMTP credentials em `.env`
4. ✅ Testar: `telnet smtp.sendgrid.net 587`

### "Authentication failed"
```bash
# Verificar API key está correta
grep "SMTP_PASS" .env

# Deve começar com: SG.
# Deve ter caracteres especiais (-._)
# Não deve ter espaços
```

### Limite de Taxa (Rate Limit)
- SendGrid permite ~100 emails/segundo
- Para seu projeto: sem problema
- Se precisar de mais: contate SendGrid

---

## 📧 Templates Funcionando

Todos os 4 templates funcionam perfeItamente:

1. **pet_registered** ✅
2. **adoption_request** ✅
3. **adoption_confirmed** ✅
4. **contact** ✅

---

## 🎯 Checklist Final

- [ ] `.env` com SendGrid API key
- [ ] `go get gopkg.in/gomail.v2` executado
- [ ] `go run main.go` inicia sem erros
- [ ] GET /api/v1/test/email-config retorna ok
- [ ] POST /api/v1/test/email envia email
- [ ] Email recebido em sua caixa
- [ ] Criar pedido de adoção dispara email
- [ ] Dashboard SendGrid mostra emails enviados

---

## 🚀 Status Final

```
╔═════════════════════════════════════════════╗
║        ✅ SENDGRID CONFIGURADO              ║
║                                             ║
║    Email Provider: SendGrid              ║
║    Método: SMTP                            ║
║    From: amigofieladocao@gmail.com        ║
║                                             ║
║    Status: 🟢 PRONTO PARA ENVIO             ║
╚═════════════════════════════════════════════╝
```

---

## 📚 Recursos

- [SendGrid SMTP Documentation](https://sendgrid.com/docs/for-developers/sending-email/getting-started-smtp/)
- [Gomail Documentation](https://pkg.go.dev/gopkg.in/gomail.v2)
- [SendGrid API Docs](https://docs.sendgrid.com/api-reference)

---

**Última atualização:** 29 de Março de 2026
**Versão:** 1.0 - SendGrid Ready
