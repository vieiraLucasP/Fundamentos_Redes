#!/bin/bash
# Script para executar uma máquina com redirecionamento de logs

if [ $# -lt 1 ]; then
    echo "Uso: $0 <arquivo_config>"
    exit 1
fi

CONFIG_FILE=$1
MACHINE_NAME=$(grep -v '^#' $CONFIG_FILE | sed -n '2p')
LOG_FILE="${MACHINE_NAME,,}_log.txt"

echo "Iniciando máquina $MACHINE_NAME com logs redirecionados para $LOG_FILE"

# Executar a máquina com redirecionamento de stderr para o arquivo de log
./bin/machine $CONFIG_FILE 2> $LOG_FILE