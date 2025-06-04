#!/bin/bash
# Script para testar comunicação UDP básica

# Uso: ./teste_udp.sh [modo] [ip_destino]
# modo: "servidor" ou "cliente"
# ip_destino: IP da máquina destino (apenas para modo cliente)

if [ "$1" == "servidor" ]; then
  echo "Iniciando servidor UDP na porta 6000..."
  echo "Pressione Ctrl+C para sair"
  nc -u -l 6000
elif [ "$1" == "cliente" ]; then
  if [ -z "$2" ]; then
    echo "Erro: Especifique o IP de destino"
    echo "Uso: ./teste_udp.sh cliente IP_DESTINO"
    exit 1
  fi
  echo "Enviando mensagem para $2:6000..."
  echo "TESTE_UDP_$(date)" | nc -u -w 1 $2 6000
  echo "Mensagem enviada!"
else
  echo "Uso: ./teste_udp.sh [servidor|cliente] [ip_destino]"
  echo "Exemplos:"
  echo "  ./teste_udp.sh servidor"
  echo "  ./teste_udp.sh cliente 192.168.0.50"
fi