# Configuração de Email com Gmail

## Passo 1: Ativar autenticação em dois fatores no Gmail
1. Acesse https://myaccount.google.com/
2. Clique em "Segurança" na barra lateral
3. Ative a Autenticação em duas etapas

## Passo 2: Gerar senha de aplicativo
1. Volte para a página de Segurança
2. Procure por "Senhas de aplicativo" (aparece após ativar 2FA)
3. Selecione "Mail" e "Windows Computer" (ou seu SO)
4. Google irá gerar uma senha de 16 caracteres
5. Copie essa senha

## Passo 3: Configurar variáveis de ambiente
Atualize o arquivo `.env` com:

```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=gustavodl.gdl33@gmail.com
SMTP_PASS=[sua_senha_de_16_caracteres_gerada]
SMTP_FROM=gustavodl.gdl33@gmail.com
```

## Passo 4: Testar a configuração
Você pode testar o envio de email completando uma solicitação de adoção através da API:

```bash
curl -X POST http://localhost:8080/api/v1/pedidos-adocao \
  -H "Content-Type: application/json" \
  -d '{
    "ong_id": "seu_ong_id",
    "pet_id": "seu_pet_id",
    "respostas": [
      {
        "campo_formulario_id": "campo_id",
        "valor": "Seu Nome"
      }
    ]
  }'
```

## Dicas importantes
- ⚠️ Nunca compartilhe sua senha de aplicativo no GitHub
- A senha de aplicativo é diferente da sua senha regular do Gmail
- Se receber erros de autenticação, regere a senha de aplicativo
- O SMTP_PORT 587 usa TLS (protocolo seguro recomendado)

## Emails configurados no sistema

### Quando um pet é registrado
- Template: `NewPetRegisteredEmail()`
- Função: `SendEmailPetRegistered()`

### Quando uma solicitação de adoção é criada
- Template: `NewAdoptionRequestEmail()`
- Função: `SendEmailAdoptionRequest()`
- Enviado para o email da ONG automaticamente via goroutine

### Quando uma adoção é confirmada
- Template: `NewAdoptionConfirmedEmail()`
- Função: `SendEmailAdoptionConfirmed()`

### Para contatos gerais
- Template: `NewContactEmail()`
- Função: `SendEmailContact()`

## Arquivos modificados

1. **utils/email.go** - Objeto Mailer que gerencia conexão SMTP
2. **utils/templates.go** - Templates de emails em HTML
3. **utils/mailer_service.go** - Serviço centralizado de envio de emails
4. **controllers/v1/pedidoadocao.go** - Integração de envio de email ao criar pedido
5. **main.go** - Inicialização do mailer no startup da aplicação
6. **.env** - Configurações SMTP
