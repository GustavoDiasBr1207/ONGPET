#!/bin/bash
# 🚀 Script para iniciar OngPet Backend com Evolution API

set -e  # Exit on error

cd "$(dirname "$0")"

echo "════════════════════════════════════════════════════════"
echo "🚀 Iniciando OngPet com Evolution API"
echo "════════════════════════════════════════════════════════"
echo ""

# Verificar Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker não encontrado. Instale em: https://docs.docker.com/get-docker/"
    exit 1
fi

# Verificar Docker Compose
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    echo "❌ Docker Compose não encontrado."
    echo "   Instale: https://docs.docker.com/compose/install/"
    exit 1
fi

echo "✅ Docker encontrado: $(docker --version)"
echo "✅ Docker Compose encontrado: $COMPOSE_CMD"
echo ""

# Parar containers antigos
echo "🛑 Parando containers antigos..."
$COMPOSE_CMD down --volumes 2>/dev/null || true
sleep 2

echo ""
echo "🔨 Compilando e iniciando containers..."
echo "   (Primeira execução pode demorar 1-2 minutos)"
echo ""

# Iniciar containers
$COMPOSE_CMD up --build

# Se chegou aqui, tudo correu bem
echo ""
echo "════════════════════════════════════════════════════════"
echo "✅ OngPet Backend está rodando!"
echo "════════════════════════════════════════════════════════"
echo ""
echo "📍 Endpoints:"
echo "   - API:           http://localhost:8081"
echo "   - Evolution API: http://localhost:8080"
echo "   - Swagger:       http://localhost:8081/swagger/index.html"
echo ""
echo "💡 Dica: Use Ctrl+C para parar"
echo ""
