# Simulação de Rede Local em Anel

Este projeto implementa uma simulação de rede local em anel usando o protocolo UDP em Go. A aplicação simula o funcionamento de uma rede em anel com passagem de token, fila de mensagens e controle de erro usando CRC32.

## Características

- **Protocolo de Comunicação**: UDP
- **Arquitetura**: Rede em anel com passagem de token
- **Controle de Erro**: CRC32
- **Fila de Mensagens**: Máximo 10 mensagens por máquina
- **Tipos de Transmissão**: Unicast e Broadcast
- **Detecção de Falhas**: Módulo de inserção de erros aleatórios

## Estrutura do Projeto

```
.
├── cmd/machine/           # Aplicação principal
├── pkg/
│   ├── config/           # Configuração da máquina
│   ├── crc/              # Cálculo de CRC32
│   ├── message/          # Tipos de mensagens e pacotes
│   └── network/          # Lógica principal da rede
├── internal/queue/       # Fila de mensagens
└── configs/              # Arquivos de configuração de exemplo
```

## Tipos de Pacotes

### Token
- Formato: `1000`
- Usado para controlar o acesso ao meio de transmissão

### Pacote de Dados
- Formato: `2000;<origem>:<destino>:<controle>:<CRC>:<mensagem>`
- Estados de controle:
  - `maquinanaoexiste`: Máquina destino não encontrada
  - `ACK`: Mensagem recebida corretamente
  - `NAK`: Erro detectado na mensagem

## Configuração

Cada máquina deve ter um arquivo de configuração com o seguinte formato:

```
<ip_destino_do_token>:porta
<apelido_da_máquina_atual>
<tempo_token>
<gera_token_inicial>
```

### Exemplo (config_alice.txt):
```
localhost:6001
Alice
2
true
```

## Compilação e Execução

### 1. Compilar o projeto
```bash
go build -o bin/machine cmd/machine/main.go
```

### 2. Executar uma máquina
```bash
go run cmd/machine/main.go config_alice.txt
```

### 3. Executar múltiplas máquinas (em terminais separados)
```bash
# Terminal 1 - Alice (gera token inicial)
go run cmd/machine/main.go config_alice.txt

# Terminal 2 - Bob
go run cmd/machine/main.go config_bob.txt

# Terminal 3 - Carol
go run cmd/machine/main.go config_carol.txt
```

## Comandos Disponíveis

Durante a execução, você pode usar os seguintes comandos:

- `send <destino> <mensagem>` - Enviar mensagem unicast
- `broadcast <mensagem>` - Enviar mensagem broadcast (para TODOS)
- `status` - Ver status da máquina
- `queue` - Ver fila de mensagens
- `token` - Gerar novo token manualmente
- `logs` - Ver últimas linhas do arquivo de log
- `help` - Mostrar comandos disponíveis
- `quit` - Sair da aplicação

### Exemplos de uso:
```
> send Bob Olá Bob, como vai?
> broadcast Olá pessoal!
> status
> queue
```

## Funcionamento

### 1. Inicialização
- Uma máquina (configurada com `true`) gera o token inicial
- O token circula pela rede em ordem (Alice → Bob → Carol → Alice)

### 2. Transmissão de Mensagens
- Máquinas só podem transmitir quando possuem o token
- Se a fila estiver vazia, o token é passado adiante
- Se há mensagens na fila, a primeira é enviada

### 3. Controle de Erro
- CRC32 é calculado para cada mensagem
- Módulo de falhas introduz erros aleatoriamente (10% de probabilidade)
- Mensagens com erro são retransmitidas uma vez

### 4. Estados de Retorno
- **ACK**: Mensagem recebida corretamente, remove da fila
- **NAK**: Erro detectado, mantém na fila para retransmissão
- **maquinanaoexiste**: Destino não encontrado, remove da fila

## Arquivos de Configuração de Exemplo

O projeto inclui três arquivos de configuração para teste:

- `config_alice.txt` - Porta 6000, gera token inicial
- `config_bob.txt` - Porta 6001
- `config_carol.txt` - Porta 6002

## Monitoramento e Depuração

A aplicação fornece logs detalhados sobre:
- Recebimento e envio de tokens
- Transmissão de mensagens
- Detecção de erros
- Status da fila de mensagens
- Atividade da rede

Os logs são gravados em arquivos de texto separados para cada máquina (ex: alice_log.txt, bob_log.txt), mantendo o terminal limpo para comandos. Use o comando `logs` para visualizar as últimas linhas do arquivo de log.

## Requisitos

- Go 1.19 ou superior
- Portas UDP 6000-6002 disponíveis para teste local

## Limitações

- Máximo 10 mensagens na fila por máquina
- Retransmissão limitada a uma tentativa por mensagem
- Configuração manual do anel (ordem das máquinas)

## Scripts de Demonstração

O projeto inclui vários scripts para facilitar o uso:

- `demo_rapida.sh` - Demonstração automatizada rápida
- `demo.sh` - Demonstração interativa completa
- `manual_test.sh` - Teste manual em terminais separados
- `test_final_validation.sh` - Validação automática completa

### Uso Rápido
```bash
# Demonstração automática
./demo_rapida.sh

# Demonstração interativa
./demo.sh

# Validação completa
./test_final_validation.sh
```

## Status do Projeto

✅ **PROJETO CONCLUÍDO COM SUCESSO**

Todas as funcionalidades foram implementadas e testadas:
- Sistema de passagem de token funcionando
- Comunicação UDP entre máquinas
- Fila de mensagens com limite de 10
- Controle de erro CRC32
- Estados ACK/NAK/maquinanaoexiste
- Transmissão unicast e broadcast
- Módulo de inserção de falhas
- Interface interativa completa
- Detecção e regeneração de token perdido

## Desenvolvimento

Para contribuir com o projeto:

1. Clone o repositório
2. Execute `go mod tidy` para baixar dependências
3. Execute os testes com `go test ./...`
4. Compile com `go build cmd/machine/main.go`

## Licença

Este projeto foi desenvolvido para fins educacionais como parte do trabalho final da disciplina de Redes de Computadores.
