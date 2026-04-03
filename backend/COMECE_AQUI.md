# 🚀 **ONGPET Backend - Guia Rápido**

## ✅ **Status Atual**

- ✅ WhatsApp **desabilitado** (funciona depois)
- ✅ API principal **pronta**
- ✅ Porta: **8081**

---

## 🎯 **Iniciando Agora**

```bash
cd /home/gustavodias/projetosvscode/ONGPET/backend

# Parar tudo
docker compose down

# Iniciar apenas API
docker compose up --build
```

---

## 📍 **Acessar a API**

```
http://localhost:8081
```

### Testar Health Check
```bash
curl http://localhost:8081/api/v1/test/email-config
```

### Swagger (Documentação)
```
http://localhost:8081/swagger/index.html
```

---

## 🔧 **Depois - Habilitar WhatsApp**

Quando Evolution API tiver funcionando:

### 1️⃣ Descomentar em `docker-compose.yml`
```yaml
# Remover # de:
  evolution:
    image: evoapicloud/evolution-api:latest
    ...
```

### 2️⃣ Reabilitar em `.env`
```env
WHATSAPP_ENABLED=true
```

### 3️⃣ Conectar WhatsApp
```bash
# Rodar tudo
docker compose down
docker compose up --build

# Acessar
http://localhost:8080
# Escanear QR code
```

---

## 📧 **Testar Email (já funciona)**

```bash
curl -X POST http://localhost:8081/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu_email@gmail.com",
    "nome": "Teste"
  }'
```

---

## 🧪 **Próximos Passos**

1. ✅ Rodar API com `docker compose up --build`
2. ✅ Acessar http://localhost:8081
3. ✅ Testar endpoints via Swagger
4. ✅ Criar ONGs, Pets, Formulários
5. ✅ Depois: Habilitar WhatsApp + Evolution

---

## 📚 **Documentação**

- [WHATSAPP_SETUP.md](WHATSAPP_SETUP.md) - WhatsApp detalhado
- [DOCKER_RUN.md](DOCKER_RUN.md) - Docker explicado
- [EVOLUTION_CORRIGIDO.md](EVOLUTION_CORRIGIDO.md) - Evolution API

---

**Está pronto para rodar! 🚀**
