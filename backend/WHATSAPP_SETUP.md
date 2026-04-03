# Integração WhatsApp com Revolution API

## 📲 Visão Geral

O sistema agora suporta envio de mensagens WhatsApp através da **Revolution API** (ou Evolution API) em paralelo com emails nas seguintes situações:

1. **Novo Pedido de Adoção**: Notifica ONG + Confirmação para solicitante
2. **Pedido Aprovado**: Envia mensagem de aprovação com contato da ONG
3. **Pedido Rejeitado**: Envia mensagem de rejeição

## ⚙️ Configuração

### Variáveis de Ambiente

Adicione as seguintes variáveis ao seu arquivo `.env`:

```env
# WhatsApp via Revolution/Evolution API
WHATSAPP_ENABLED=true
WHATSAPP_API_URL=http://localhost:3000/api/messages/send  # ou sua URL da API
WHATSAPP_API_KEY=seu_token_api_aqui
```

### Valores Esperados

| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `WHATSAPP_ENABLED` | Habilita/desabilita o envio de WhatsApp | `true` ou `false` |
| `WHATSAPP_API_URL` | URL base da API Revolution/Evolution | `http://localhost:3000/api/messages/send` |
| `WHATSAPP_API_KEY` | Token de autenticação da API | `seu_token_secreto` |

### Dependências de Configuração

Para que o sistema funcione corretamente, você precisa:

1. **Banco de Dados com campos obrigatórios:**
   - Campo `Telefone` na tabela `Ong` (já existe)
   - Campo `Telefone` obrigatório (`Obrigatorio: true`) no formulário dinâmico do Pet

2. **Dados Corretos:**
   - Cada ONG deve ter seu telefone preenchido: `Ong.Telefone` (formato: 55XXXXXXXXXX)
   - O formulário de adoção deve ter um campo "Telefone" obrigatório
   - Os números devem estar no formato: **55 + DDD + Número** (ex: 5527992345678)

## 🔧 Setup da Revolution/Evolution API

A Revolution API é uma solução não-oficial que automatiza o WhatsApp Web. Para testá-la:

### Opção 1: Usar Docker (Recomendado)

```bash
docker pull atendai/evolution-api:latest
docker run -p 3000:3000 atendai/evolution-api:latest
```

### Opção 2: Instalação Manual

1. Clone o repositório: `git clone https://github.com/EvolutionAPI/evolution-api.git`
2. Instale as dependências: `npm install`
3. Configure as variáveis de ambiente
4. Inicie o servidor: `npm start`

### Opção 3: Serviço Cloud

Use um serviço gerenciado de Evolution API (existem várias opções comerciais)

## 📋 Estrutura das Mensagens

### 1. Novo Pedido → ONG

```
🐾 *Nova Solicitação de Adoção!*

Olá {ONG},

Você recebeu uma nova solicitação de adoção para:
🐶 *{PET}*

👤 *Solicitante:* {NOME}

Acesse o painel da ONG para revisar...
```

### 2. Confirmação → Solicitante

```
🎉 *Solicitação Recebida!*

Olá {NOME},

Sua solicitação de adoção para *{PET}* foi recebida com sucesso! ✅

A *{ONG}* está analisando seu pedido...
```

### 3. Aprovado → Solicitante

```
🎉 *PARABÉNS! Sua Adoção foi Aprovada!*

Olá {NOME},

Sua solicitação para *{PET}* foi *APROVADA*! ✅

📞 Entre em contato com a ONG:
📱 {TELEFONE_ONG}
```

### 4. Rejeitado → Solicitante

```
😔 *Solicitação não aprovada*

Olá {NOME},

Infelizmente, sua solicitação para *{PET}* não foi aprovada...
```

## 📊 Fluxo de Dados

```
Criar Pedido de Adoção
    ↓
[Email Enviado para ONG] + [WhatsApp Enviado para ONG + Solicitante]
    ↓
ONG Aprecia/Rejeita
    ↓
[Email Enviado para Solicitante] + [WhatsApp Enviado para Solicitante]
```

## 🚀 Logging e Monitoramento

Quando habilitado, o sistema registra:

- ✅ Sucesso: `✉️ WhatsApp enviado (tipo) para ****1234 em 234ms`
- ⚠️ Erro: `⚠️ Erro ao enviar WhatsApp para ****1234 (após 2 tentativas): motivo`

Os números de telefone são mascarados no log por segurança (mostra apenas últimos 4 dígitos).

## 🛡️ Tratamento de Erros

O sistema implementa:

1. **Graceful Degradation**: Se WhatsApp está desabilitado, não causa erros
2. **Retry com Backoff**: Máximo 2 tentativas com delay exponencial (1s, 2s)
3. **Não-bloqueante**: Falhas de WhatsApp não interrompem o fluxo de criação/atualização de pedidos
4. **Logging**: Todos os erros são registrados para debug

## 🧪 Testando a Integração

### 1. Verificar Inicialização

Ao iniciar o servidor, você deverá ver:

```
✅ WhatsApp service inicializado com sucesso
```

Ou se desabilitado:

```
ℹ️ WhatsApp service desabilitado (WHATSAPP_ENABLED != true)
```

### 2. Testar Manualmente

```bash
# Ver logs
docker logs -f seu_container_whatsapp

# Ou testar via curl
curl -X POST http://localhost:3000/api/messages/send \
  -H "Authorization: Bearer seu_token" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "5527992345678",
    "body": "Teste de mensagem"
  }'
```

### 3. Verificar no Banco

```sql
-- Confirmar que ONG tem telefone
SELECT nome, telefone FROM ong WHERE telefone IS NOT NULL;

-- Confirmar que formulário tem campo de telefone obrigatório
SELECT nome, formulario_id FROM pet WHERE formulario_id IS NOT NULL;
```

## ⚡ Performance

- Cada mensagem é enviada em **goroutine assíncrona** (não bloqueia)
- Timeout de conexão: **10 segundos**
- Retry automático: até **2 tentativas**

## 🔐 Segurança

- API Key deve estar em `WHATSAPP_API_KEY` no `.env`
- Números de telefone são validados antes do envio
- URLs são construídas dinamicamente sem string concatenation perigosa
- Logs masscaram números de telefone (apenas últimos 4 dígitos)

## 📚 Referências

- [Evolution API Docs](https://github.com/EvolutionAPI/evolution-api)
- [Formato de Números WhatsApp](https://developers.facebook.com/docs/whatsapp/phone-numbers/)
- [Go HTTP Client](https://golang.org/pkg/net/http/)

## 🆘 Troubleshooting

| Problema | Solução |
|----------|---------|
| "WHATSAPP_API_URL não configurado" | Adicionar `WHATSAPP_API_URL` no `.env` |
| "Número inválido" | Usar formato 55XXXXXXXXXX (sem caracteres especiais) |
| "API retornou status 401" | Verificar `WHATSAPP_API_KEY` |
| "Timeout" | Aumentar timeout (padrão 10s) ou verificar conexão com API |
| "Desabilitado" | Mudar `WHATSAPP_ENABLED=true` |

---

## ✅ Checklist de Implementação

- [x] Criar `whatsapp_service.go` com SDK da API
- [x] Criar `whatsapp_messages.go` com templates
- [x] Criar `whatsapp_sender.go` com funções de disparo
- [x] Modificar `CreatePedidoAdocao` para enviar WhatsApp
- [x] Modificar `UpdateStatusPedidoAdocao` para enviar WhatsApp
- [x] Inicializar serviço no `main.go`
- [ ] Testar fluxo completo end-to-end
- [ ] Configurar variáveis de ambiente em produção
- [ ] Documentar endpoints de teste (opcional)
