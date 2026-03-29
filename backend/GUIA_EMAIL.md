# 📧 Sistema de Email - Guia Completo

## ✅ O que foi completado

A lógica de envio de email foi implementada com sucesso para o seu projeto ONGPET. Aqui está tudo o que foi criado e modificado:

### Arquivos Criados
- ✨ `utils/mailer_service.go` - Serviço centralizado de envio
- ✨ `utils/email_test.go` - Utilitários para testes
- ✨ `controllers/v1/email_test.go` - Endpoints de teste
- 📄 `EMAIL_CONFIG.md` - Documentação de configuração

### Arquivos Modificados
- 🔄 `utils/email.go` - Corrigido package name (mailer → utils)
- 🔄 `utils/templates.go` - Adicionados 3 novos templates de email
- 🔄 `controllers/v1/pedidoadocao.go` - Integrado envio de email automático
- 🔄 `controllers/routes.go` - Adicionadas rotas de teste
- 🔄 `main.go` - Inicialização do mailer ao startup
- 🔄 `.env` - Configurado com seu email: gustavodl.gdl33@gmail.com

## 🚀 Próximos passos

### 1. Instalar dependência do Gomail

```bash
cd /home/gustavodias/projetosvscode/ONGPET/backend
go get -u gopkg.in/gomail.v2
go mod tidy
```

### 2. Configurar Gmail

Siga os passos em [EMAIL_CONFIG.md](./EMAIL_CONFIG.md):

1. Ativar 2FA no Gmail (https://myaccount.google.com/)
2. Gerar senha de aplicativo em "Senhas de aplicativo"
3. Atualizar `.env` com a senha de 16 caracteres:
   ```
   SMTP_PASS=sua_senha_de_16_caracteres
   ```

### 3. Testar a configuração

#### Verificar status do email
```bash
curl http://localhost:8080/api/v1/test/email-config
```

Resposta esperada:
```json
{
  "status": "ok",
  "message": "Email configurado corretamente",
  "from": "gustavodl.gdl33@gmail.com"
}
```

#### Enviar email de teste
```bash
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "adoption_request",
    "data": {
      "pet_name": "Rex",
      "requester_name": "João Silva",
      "ong_name": "ONG Paws"
    }
  }'
```

## 📧 Templates de Email Disponíveis

### 1. Pet Registered
Enviado quando um pet é registrado
```json
{
  "template": "pet_registered",
  "data": {
    "pet_name": "Fluffy",
    "owner_name": "Maria"
  }
}
```

### 2. Adoption Request
Enviado automaticamente quando uma solicitação de adoção é criada
```json
{
  "template": "adoption_request",
  "data": {
    "pet_name": "Buddy",
    "requester_name": "João",
    "ong_name": "ONG Pets"
  }
}
```

### 3. Adoption Confirmed
Enviado quando uma adoção é aprovada
```json
{
  "template": "adoption_confirmed",
  "data": {
    "pet_name": "Max",
    "requester_name": "Ana"
  }
}
```

### 4. Contact
Para contatos gerais
```json
{
  "template": "contact",
  "data": {
    "name": "Pedro",
    "email": "pedro@email.com",
    "message": "Olá, gostaria de adotar..."
  }
}
```

## 🔄 Integração Automática

### Quando criar pedido de adoção
A API enviará automaticamente um email para a ONG:

```bash
POST /api/v1/pedidos-adocao
{
  "ong_id": "uuid-da-ong",
  "pet_id": "uuid-do-pet",
  "respostas": [
    {
      "campo_formulario_id": "uuid-campo",
      "valor": "Seu Nome"
    }
  ]
}
```

✅ Email será enviado via goroutine (não bloqueia a resposta)

## 🛠️ Troubleshooting

### Email não enviado?
1. Certifique-se de ter instalado: `go get gopkg.in/gomail.v2`
2. Verifique as variáveis SMTP no `.env`
3. Teste com: `curl http://localhost:8080/api/v1/test/email-config`
4. Verifique os logs da aplicação

### Erro "authentication failed"
- Regere a senha de aplicativo do Gmail
- Certifique-se que há espaço em branco no .env
- Verificar se a senha tem exatamente 16 caracteres

### "Connection refused"
- Verificar se SMTP_HOST é smtp.gmail.com
- Verificar SMTP_PORT é 587
- Ter conexão com internet

## 📝 Notas Importantes

- ⚠️ A senha de aplicativo é diferente da senha regular do Gmail
- 🔒 Nunca compartilhe a senha no GitHub (use .env.local em produção)
- 🚀 Os emails são enviados em goroutines para não bloquear requisições
- 📧 Todos os emails usam HTML e podem incluir imagens/links

## 📚 Documentação

Para mais detalhes, consulte:
- [EMAIL_CONFIG.md](./EMAIL_CONFIG.md) - Configuração Gmail
- `utils/email.go` - Implementação do Mailer
- `utils/templates.go` - Templates HTML
- `utils/mailer_service.go` - Serviço de emails
