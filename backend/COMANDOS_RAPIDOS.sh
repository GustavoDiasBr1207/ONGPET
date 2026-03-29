#!/usr/bin/env bash
# COPY-PASTE RÁPIDO - Comandos para implementação de Email
# ═══════════════════════════════════════════════════════════════

# 1️⃣  INSTALAR DEPENDÊNCIA
# ───────────────────────────────────────────────────────────────
cd /home/gustavodias/projetosvscode/ONGPET/backend
go get -u gopkg.in/gomail.v2
go mod tidy


# 2️⃣  INICIAR SERVIDOR
# ───────────────────────────────────────────────────────────────
go run main.go


# 3️⃣  VERIFICAR CONFIGURAÇÃO (em outro terminal)
# ───────────────────────────────────────────────────────────────
curl http://localhost:8080/api/v1/test/email-config


# 4️⃣  ENVIAR EMAIL DE TESTE
# ───────────────────────────────────────────────────────────────
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "adoption_request",
    "data": {
      "pet_name": "Buddy",
      "requester_name": "João Silva",
      "ong_name": "ONG Pets"
    }
  }'


# 5️⃣  ENVIAR EMAIL COM OUTRO TEMPLATE
# ───────────────────────────────────────────────────────────────

# Pet Registered
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "pet_registered",
    "data": {
      "pet_name": "Rex",
      "owner_name": "Maria"
    }
  }'

# Adoption Confirmed
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "adoption_confirmed",
    "data": {
      "pet_name": "Fluffy",
      "requester_name": "Ana Silva"
    }
  }'

# Contact
curl -X POST http://localhost:8080/api/v1/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "gustavodl.gdl33@gmail.com",
    "template": "contact",
    "data": {
      "name": "Pedro",
      "email": "pedro@email.com",
      "message": "Olá, gostaria de adotar um pet..."
    }
  }'


# 6️⃣  EXECUTAR SCRIPT DE TESTE AUTOMÁTICO
# ───────────────────────────────────────────────────────────────
chmod +x test-email.sh
./test-email.sh http://localhost:8080 gustavodl.gdl33@gmail.com


# 7️⃣  LER LOGS DA APLICAÇÃO (em tempo real)
# ───────────────────────────────────────────────────────────────
# A saída do `go run main.go` mostrará:
# ✅ Mailer inicializado com sucesso
# ✅ Email enviado para ONG: email@example.com
# ⚠️ Erro ao enviar email: [motivo]


# 8️⃣  CRIAR PEDIDO DE ADOÇÃO (Teste completo)
# ───────────────────────────────────────────────────────────────
# Substitua os UUIDs pelos valores reais do seu banco de dados

curl -X POST http://localhost:8080/api/v1/pedidos-adocao \
  -H "Content-Type: application/json" \
  -d '{
    "ong_id": "11111111-1111-1111-1111-111111111111",
    "pet_id": "22222222-2222-2222-2222-222222222222",
    "respostas": [
      {
        "campo_formulario_id": "33333333-3333-3333-3333-333333333333",
        "valor": "João da Silva"
      },
      {
        "campo_formulario_id": "44444444-4444-4444-4444-444444444444",
        "valor": "joao@email.com"
      }
    ]
  }'

# Esperado:
# HTTP 201 Created (imediato)
# + Email enviado para ONG em background


# 9️⃣  VERIFICAR COMPILAÇÃO SEM ERROS
# ───────────────────────────────────────────────────────────────
go build -o tmp/main main.go
echo "Compilação OK: $?"


# 🔟 LIMPAR ARQUIVOS TEMPORÁRIOS
# ───────────────────────────────────────────────────────────────
rm -f tmp/main
go clean


# ⚠️  EDITAR .env PARA ADICIONAR SENHA GMAIL
# ───────────────────────────────────────────────────────────────

# 1. Gere sua senha no Gmail: https://myaccount.google.com/
# 2. Edite o arquivo .env:

nano .env
# ou
vim .env
# ou
code .env  # no VS Code

# Procure por:
# SMTP_PASS=sua_senha_de_app

# E substitua por sua senha de 16 caracteres
# Exemplo:
# SMTP_PASS=ffxz mtvb rbmv wzaq
#          (sem espaços, copie como está do Gmail)


# ✅ APÓS COMPLETAR TUDO
# ───────────────────────────────────────────────────────────────

echo "✅ Sistema de Email Configurado!"
echo "📧 Email: gustavodl.gdl33@gmail.com"
echo "🚀 Pronto para enviar!"


# 📚 LEIA A DOCUMENTAÇÃO
# ───────────────────────────────────────────────────────────────
# cat RESUMO_EMAIL.md
# cat GUIA_EMAIL.md
# cat EMAIL_CONFIG.md
# cat ARQUITETURA_EMAIL.md
# cat CHECKLIST_IMPLEMENTACAO.md
