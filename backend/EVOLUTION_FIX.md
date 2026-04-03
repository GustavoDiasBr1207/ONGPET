# ✅ CORREÇÔES APLICADAS - Evolution API

## 🔴 Problema Original

```
evolution_api exited with code 1 (restarting)
```

**Causa**: Imagem `atendai/evolution-api` não encontrada ou incompatível.

---

## ✅ Soluções Implementadas

### 1️⃣ docker-compose.yml Atualizado

**Antes (❌ Errado)**:
```yaml
evolution:
  image: atendai/evolution-api
  container_name: evolution_api
```

**Depois (✅ Correto)**:
```yaml
evolution:
  image: ghcr.io/evolutionapi/evolution-api:latest
  container_name: evolution_api
  environment:
    - AUTHENTICATION_API_KEY=123456
    - DATABASE_ENABLED=false
    - DATABASE_SAVE_DATA_ON_MEMORY=true
    - SERVER_PORT=8080
    - LOGGER_LEVEL=info
    # ... (outras variáveis)
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 40s
```

**Melhorias**:
✅ Imagem oficial do repositório GitHub (`ghcr.io`)  
✅ Variáveis de ambiente necessárias configuradas  
✅ Health check para monitorar status  
✅ Startup period de 40s (Evolution API demora para inicializar)  

### 2️⃣ .env Atualizado

**Antes**:
```env
WHATSAPP_API_URL=http://evolution:8080/message/sendText/ongpet
```

**Depois**:
```env
WHATSAPP_API_URL=http://evolution:8080/message/sendText
WHATSAPP_API_KEY=123456
```

### 3️⃣ whatsapp_service.go Atualizado

**Compatibilidade com Evolution API Oficial**:

```go
// Payload format
{
  "number": "5527992345678",
  "text": "Sua mensagem aqui"
}

// Headers
Content-Type: application/json
Authorization: Bearer 123456
```

---

## 🚀 Como Rodar Agora

### Via Docker Desktop/CLI

```bash
cd /home/gustavodias/projetosvscode/ONGPET/backend

# Opção 1: Docker Compose (nova sintaxe)
docker compose up --build

# Opção 2: docker-compose (sintaxe antiga)
docker-compose up --build

# Detached (background)
docker compose up -d --build
```

### Esperado:

```log
evolution_api      | info: Server running on port 8080
ongpet_app         | ✅ WhatsApp service inicializado com sucesso
ongpet_app         | 🚀 API rodando em http://localhost:8081
```

### Portas

| Serviço | Porta | Status |
|---------|-------|--------|
| OngPet API | 8081 | `/api/v1/health` |
| Evolution API | 8080 | health check automático |

---

## ⏱️ Timeline de Inicialização

```
0s    - Containers estão sendo iniciados
15s   - Evolution API compilando inicialização
30s   - Database preparando (em memória, rápido)
40s   - Health check validando
45s+  - OngPet se conecta à Evolution API
50s   - Tudo pronto! 🎉
```

**Se Evolution falhar após 50s**, verificar:
```bash
docker logs evolution_api
docker logs ongpet_app
```

---

## 📝 Próximas Etapas

1. ✅ Rodar `docker compose up --build`
2. ✅ Aguardar 50 segundos (primeira execução é lenta)
3. ✅ Verificar logs: `docker logs -f evolution_api`
4. ✅ Testar: `curl http://localhost:8080`
5. ✅ Criar pedido de adoção com telefone
6. ✅ Verfir WhatsApp sendo enviado nos logs

---

## 🔗 Referências

- [Evolution API Official](https://github.com/EvolutionAPI/evolution-api)
- [DOCKER_RUN.md](DOCKER_RUN.md) - Guia completo
- [WHATSAPP_SETUP.md](WHATSAPP_SETUP.md) - Configuração WhatsApp

---

## ✅ Checklist Pré-execução

- [ ] Docker instalado: `docker --version`
- [ ] Docker Compose disponível: `docker compose --version` ou `docker-compose --version`
- [ ] `.env` configurado com credenciais do banco
- [ ] Nenhum container rodando na porta 8080 ou 8081
- [ ] Pelo menos 2GB de RAM livre

---

**Status**: ✅ Pronto para segunda execução
