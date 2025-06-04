#!/bin/bash

# 🔧 SCRIPT PARA PREPARAR SUA MÁQUINA PARA REDE FÍSICA
# ====================================================

echo "🌐 PREPARANDO SUA MÁQUINA PARA REDE EM ANEL"
echo "============================================"
echo

# 1. Descobrir IP local
echo "1️⃣  DESCOBRINDO SEU IP NA REDE WI-FI..."
MY_IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | head -1 | awk '{print $2}')

if [ -z "$MY_IP" ]; then
    echo "❌ Não foi possível descobrir seu IP automaticamente"
    echo "   Execute manualmente: ifconfig | grep inet"
    echo "   Anote seu IP da rede Wi-Fi (exemplo: 192.168.1.XXX)"
    exit 1
fi

echo "   ✅ Seu IP é: $MY_IP"
echo

# 2. Compilar o projeto
echo "2️⃣  COMPILANDO O PROJETO..."
if [ ! -f "bin/machine" ]; then
    echo "   Compilando..."
    go build -o bin/machine cmd/machine/main.go
    if [ $? -ne 0 ]; then
        echo "   ❌ Erro na compilação!"
        exit 1
    fi
fi
echo "   ✅ Binário pronto em bin/machine"
echo

# 3. Instruções de configuração
echo "3️⃣  CONFIGURAÇÃO NECESSÁRIA:"
echo "   📋 Combinem entre vocês:"
echo "      • Quem será Alice (gera token)"
echo "      • Quem será Bob (intermediário)"  
echo "      • Quem será Carol (fecha anel)"
echo
echo "   🔧 Cada um deve criar seu arquivo config_minha_maquina.txt:"
echo "      Linha 1: IP_DA_PRÓXIMA_MÁQUINA:porta"
echo "      Linha 2: SEU_NOME (Alice, Bob ou Carol)"
echo "      Linha 3: 3 (tempo do token)"
echo "      Linha 4: true apenas para Alice, false para outros"
echo

# 4. Exemplo prático
echo "4️⃣  EXEMPLO DE CONFIGURAÇÃO:"
echo "   Se você for ALICE e o Bob tem IP 192.168.1.105:"
echo "   ┌─────────────────────────────┐"
echo "   │ 192.168.1.105:6001         │"
echo "   │ Alice                       │"
echo "   │ 3                           │"
echo "   │ true                        │"
echo "   └─────────────────────────────┘"
echo

# 5. Testes de conectividade
echo "5️⃣  TESTE DE CONECTIVIDADE:"
echo "   📡 Peça os IPs dos colegas e teste:"
echo "      ping IP_DO_COLEGA"
echo "   🔌 Teste se as portas UDP estão abertas"
echo

echo "6️⃣  EXECUÇÃO:"
echo "   🚀 Quando todos estiverem prontos:"
echo "      ./bin/machine config_minha_maquina.txt"
echo

echo "📚 MAIS DETALHES: Veja GUIA_REDE_FISICA.md"
echo
echo "✅ PREPARAÇÃO CONCLUÍDA!"
echo "   Seu IP: $MY_IP"
echo "   Binário: bin/machine (pronto)"
echo "   Próximo passo: Criar arquivo de configuração"
