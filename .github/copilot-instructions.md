# Instruções do Copilot - Simulação de Rede em Anel

<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

Este é um projeto Go que implementa uma simulação de rede local em anel com as seguintes características:

## Contexto do Projeto
- Simulação de rede em anel usando protocolo UDP
- Sistema de passagem de token (token passing)
- Fila de mensagens para cada máquina
- Controle de erro usando CRC32
- Tipos de pacote: Token (1000) e Dados (2000)
- Suporte a transmissão Unicast e Broadcast

## Estrutura dos Pacotes
- Token: "1000"
- Dados: "2000;<origem>:<destino>:<controle>:<CRC>:<mensagem>"
- Estados de controle: "maquinanaoexiste", "ACK", "NAK"

## Funcionalidades Principais
- Geração e controle de token
- Fila de mensagens (máximo 10)
- Cálculo e verificação de CRC32
- Módulo de inserção de falhas aleatórias
- Detecção de token perdido ou duplicado
- Interface para envio de mensagens e monitoramento

## Boas Práticas
- Use goroutines para operações concorrentes
- Implemente channels para comunicação entre goroutines
- Use mutex para proteção de recursos compartilhados
- Estruture o código em packages separados para melhor organização
- Comente o código adequadamente
- Implemente logs detalhados para depuração
