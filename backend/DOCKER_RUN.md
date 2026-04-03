# 🚀 Rodando o Backend com Docker Compose (com Evolution API)

## Pré-requisitos

- Docker instalado
- Docker Compose instalado
- `.env` configurado com credenciais

## 📋 Arquivos Modificados

✅ `docker-compose.yml` — Atualizado com Evolution API oficial  
✅ `.env` — Configuração WhatsApp adicionada  
✅ `utils/whatsapp_service.go` — Compatível com Evolution API oficial  

## 🎯 Como Rodar

### 1. Limpar containers antigos (se houver)

```bash
docker-compose down
```

### 2. Iniciar os serviços

```bash
docker-compose up --build
```

Você deverá ver:

```
ongpet_app     | ✅ WhatsApp service inicializado com sucesso
ongpet_app     | 🚀 API rodando em http://localhost:8080
evolution_api  | info: Server running on port 8080
```

### 3. Portas Utilizadas

| Serviço | Porta | URL |
|---------|-------|-----|
| OngPet API | 8081 | http://localhost:8081 |
| Evolution API | 8080 | http://localhost:8080 |

## 🔧 Configuração Evolution API

A Evolution API agora está rodando com:

- **API Key**: `123456` (configure em `.env` > `WHATSAPP_API_KEY`)
- **Endpoint**: `http://evolution:8080/message/sendText`
- **Database**: Em memória (dados não persistem entre reinicializações)

### Para Persistência (Produção)

Se quiser persistência de dados, adicione MongoDB ao docker-compose:

```yaml
mongodb:
  image: mongo:latest
  container_name: mongodb
  ports:
    - "27017:27017"
  networks:
    - ongpet-network
```

E altere em `evolution`:

```yaml
DATABASE_CONNECTION_URI=mongodb://mongodb:27017
DATABASE_ENABLED=true
```

## 📱 Testando WhatsApp

### 1. Criar uma ONG com Telefone

```bash
curl -X POST http://localhost:8081/api/v1/ongsocorro \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "ONG Teste",
    "email": "ong@test.com",
    "telefone": "5527992345678",
    "endereco": "Rua Teste",
    "descricao": "ONG teste"
  }'
```

### 2. Criar um Formulário com Campo Telefone Obrigatório

O formulário deve ter:
- Campo "Nome" (obrigatório)
- Campo "Email" (obrigatório)
- Campo "Telefone" (obrigatório)

### 3. Criar um Pet com este Formulário

```bash
curl -X POST http://localhost:8081/api/v1/pets \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Rex",
    "especie": "Cachorro",
    "raca": "Vira-lata",
    "idade": 3,
    "peso": 25.5,
    "porte": "Médio",
    "ong_id": "seu_ong_id",
    "formulario_id": "seu_formulario_id"
  }'
```

### 4. Criar Pedido de Adoção

```bash
curl -X POST http://localhost:8081/api/v1/pedidos-adocao \
  -H "Content-Type: application/json" \
  -d '{
    "ong_id": "seu_ong_id",
    "pet_id": "seu_pet_id",
    "respostas": [
      {"campo_formulario_id": "campo_nome_id", "valor": "João Silva"},
      {"campo_formulario_id": "campo_email_id", "valor": "joao@example.com"},
      {"campo_formulario_id": "campo_telefone_id", "valor": "5527991234567"}
    ]
  }'
```

Você verá nos logs:

```
✉️ WhatsApp enviado (nova adoção → ONG) para ****5678 em 234ms
✉️ WhatsApp enviado (confirmação → solicitante) para ****4567 em 189ms
```

## 📊 Logs Disponíveis

Os containers estão configurados para mostrar logs estruturados:

```bash
# Ver apenas OngPet
docker-compose logs ongpet_app

# Ver apenas Evolution
docker-compose logs evolution

# Acompanhar em tempo real
docker-compose logs -f

# Últimas N linhas
docker-compose logs --tail=50
```

## 🆘 Troubleshooting

### Evolution não inicia

```
evolution_api exited with code 1
```

**Solução**: Aguarde 30-40s na primeira inicialização (healthcheck). Evolution API demora para inicializar.

```bash
docker-compose logs evolution
```

Se houver erro de EADDRINUSE:

```bash
docker-compose down
# Limpar containers específicos
docker rm evolution_api ongpet_app
docker-compose up --build
```

### WhatsApp não envia

**Verificar**:
1. ✅ `WHATSAPP_ENABLED=true` no `.env`
2. ✅ `WHATSAPP_API_URL=http://evolution:8080/message/sendText`
3. ✅ Evolution API está rodando: `docker ps | grep evolution`
4. ✅ Telefone da ONG preenchido: formato `55XXXXXXXXXX`
5. ✅ Campo "Telefone" obrigatório no formulário

**Verificar conexão com Evolution**:

```bash
curl -X GET http://localhost:8080 -v
```

### Banco de dados não conecta

**Verificar `.env`**:

```env
DATABASE_URL=postgresql://usuario:senha@host:porta/banco
```

Certifique-se que o banco está rodando antes de iniciar containers.

## 📝 Environment Variables

Consulte `.env.whatsapp.example` para variáveis completas.

Principais para WhatsApp:

```env
WHATSAPP_ENABLED=true
WHATSAPP_API_URL=http://evolution:8080/message/sendText
WHATSAPP_API_KEY=123456
```

## 🔗 Links Úteis

- [Evolution API Docs](https://github.com/EvolutionAPI/evolution-api)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)

## ✅ Checklist

- [ ] Docker e Docker Compose instalados
- [ ] `.env` configurado com credenciais corretas
- [ ] Nenhum container antigo rodando nas mesmas portas
- [ ] Primeira execução: aguarde 40s para Evolution inicializar
- [ ] Verificar logs: `docker-compose logs -f`
- [ ] API respondendo em `http://localhost:8081`
- [ ] Evolution respondendo em `http://localhost:8080`

---

**Status**: ✅ Pronto para usar
