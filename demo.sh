#!/bin/bash

# Script de demonstração da Rede em Anel
# Este script facilita a execução de múltiplas máquinas para teste

echo "=== Demonstração da Simulação de Rede em Anel ==="
echo ""
echo "Este script irá ajudar você a configurar e testar a rede em anel."
echo ""

# Verificar se o projeto está compilado
if [ ! -f "bin/machine" ]; then
    echo "Compilando projeto..."
    go build -o bin/machine cmd/machine/main.go
    if [ $? -ne 0 ]; then
        echo "Erro na compilação. Verifique o código."
        exit 1
    fi
    echo "Compilação concluída!"
    echo ""
fi

echo "Para testar a rede em anel, você precisa executar 3 máquinas em terminais separados:"
echo ""
echo "1. Terminal 1 (Alice - gera o token inicial):"
echo "   cd $(pwd)"
echo "   ./bin/machine config_alice.txt"
echo ""
echo "2. Terminal 2 (Bob):"
echo "   cd $(pwd)"
echo "   ./bin/machine config_bob.txt"
echo ""
echo "3. Terminal 3 (Carol):"
echo "   cd $(pwd)"
echo "   ./bin/machine config_carol.txt"
echo ""
echo "IMPORTANTE: Inicie Alice primeiro, depois Bob, e por último Carol."
echo "Aguarde alguns segundos entre cada inicialização."
echo ""
echo "Comandos de teste que você pode usar nas máquinas:"
echo "- send Bob Olá Bob!"
echo "- send Carol Como vai Carol?"
echo "- broadcast Olá pessoal!"
echo "- status"
echo "- queue"
echo ""
echo "Pressione qualquer tecla para continuar ou Ctrl+C para sair..."
read -n 1 -s

echo ""
echo "Escolha uma opção:"
echo "1. Executar Alice (Terminal atual)"
echo "2. Executar Bob (Terminal atual)"
echo "3. Executar Carol (Terminal atual)"
echo "4. Apenas mostrar comandos novamente"
echo "5. Sair"
echo ""
read -p "Opção (1-5): " choice

case $choice in
    1)
        echo "Executando Alice..."
        ./bin/machine config_alice.txt
        ;;
    2)
        echo "Executando Bob..."
        ./bin/machine config_bob.txt
        ;;
    3)
        echo "Executando Carol..."
        ./bin/machine config_carol.txt
        ;;
    4)
        echo ""
        echo "Comandos para executar em terminais separados:"
        echo "Terminal 1: ./bin/machine config_alice.txt"
        echo "Terminal 2: ./bin/machine config_bob.txt"
        echo "Terminal 3: ./bin/machine config_carol.txt"
        ;;
    5)
        echo "Saindo..."
        exit 0
        ;;
    *)
        echo "Opção inválida."
        exit 1
        ;;
esac
