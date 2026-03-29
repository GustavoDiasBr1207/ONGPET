#!/bin/bash

# 📧 Script de Teste de Email - OngPet
# Este script testa a configuração de email da aplicação

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         🐾 TESTE DE CONFIGURAÇÃO DE EMAIL - ONGPET        ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

API_URL="${1:-http://localhost:8080}"
TEST_EMAIL="${2:-gustavodl.gdl33@gmail.com}"

echo "📍 URL da API: $API_URL"
echo "📧 Email de teste: $TEST_EMAIL"
echo ""

# Teste 1: Verificar se API está rodando
echo "🔍 Teste 1: Verificando se a API está rodando..."
if curl -s "$API_URL/health" > /dev/null; then
    echo "✅ API está rodando"
else
    echo "❌ API não está acessível em $API_URL"
    exit 1
fi
echo ""

# Teste 2: Verificar configuração de email
echo "🔍 Teste 2: Verificando configuração de email..."
RESPONSE=$(curl -s "$API_URL/api/v1/test/email-config")
if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo "✅ Email está configurado"
    echo "   Resposta: $RESPONSE"
else
    echo "⚠️  Email pode não estar configurado"
    echo "   Resposta: $RESPONSE"
fi
echo ""

# Teste 3: Enviar email de teste
echo "🔍 Teste 3: Enviando email de teste..."
SEND_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/test/email" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "'$TEST_EMAIL'",
    "template": "adoption_request",
    "data": {
      "pet_name": "Buddy",
      "requester_name": "Teste Sistema",
      "ong_name": "ONG Teste"
    }
  }')

if echo "$SEND_RESPONSE" | grep -q '"message"'; then
    echo "✅ Email enviado com sucesso para $TEST_EMAIL"
    echo "   Resposta: $SEND_RESPONSE"
    echo ""
    echo "📧 Verifique seu email para confirmar o recebimento"
else
    echo "❌ Erro ao enviar email"
    echo "   Resposta: $SEND_RESPONSE"
    exit 1
fi
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                    ✅ TESTES CONCLUÍDOS                   ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📝 Próximas ações:"
echo "  1. Verifique seu email por uma mensagem de teste"
echo "  2. Se recebeu, o sistema está pronto para produção"
echo "  3. Se não recebeu, consulte GUIA_EMAIL.md para troubleshooting"
