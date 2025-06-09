# Relatório: Implementação de Rede em Anel com Token Ring

## 1. Estrutura da Solução

A solução implementa uma rede em anel utilizando o protocolo Token Ring, onde múltiplas máquinas se comunicam através de um token que circula pela rede. O projeto foi desenvolvido em Go (Golang) e utiliza comunicação via sockets UDP para troca de mensagens entre as máquinas.

### 1.1 Organização do Código

O projeto está organizado em uma estrutura modular:

```
/
|-- cmd/
|   `-- machine/
|       `-- main.go           # Ponto de entrada da aplicação
|-- internal/
|   `-- queue/
|       `-- queue.go          # Implementação da fila de mensagens
|-- pkg/
|   |-- config/
|   |   `-- config.go         # Configuração das máquinas
|   |-- crc/
|   |   `-- crc.go            # Implementação do CRC32
|   |-- message/
|   |   `-- message.go        # Estruturas de mensagens
|   `-- network/
|       `-- machine.go        # Implementação da máquina na rede
|-- config_teste_local.txt    # Arquivo de configuração para testes
|-- Makefile                  # Automação de compilação e execução
|-- go.mod                    # Dependências do projeto
```

### 1.2 Arquitetura

A solução implementa uma arquitetura de rede em anel (ring network) com as seguintes características:

- **Token Ring**: Protocolo onde um token circula pela rede, permitindo que apenas a máquina que possui o token possa enviar mensagens.
- **Comunicação UDP**: Utiliza sockets UDP para comunicação entre as máquinas.
- **Detecção de Erros**: Implementa verificação de integridade usando CRC32.
- **Retransmissão**: Mensagens corrompidas são retransmitidas.
- **Broadcast**: Suporte para mensagens de broadcast para todas as máquinas.

## 2. Estruturas de Dados

### 2.1 Mensagens

```go
// Mensagem na fila para envio
type QueuedMessage struct {
    Destination string     // Destino da mensagem
    Content     string     // Conteúdo da mensagem
    Timestamp   time.Time  // Momento de criação
    Retries     int        // Número de tentativas de envio
}

// Pacote de dados para transmissão na rede
type DataMessage struct {
    Type        string     // Tipo do pacote (2000 para dados)
    Origin      string     // Origem da mensagem
    Destination string     // Destino da mensagem
    Control     string     // Campo de controle (ACK, NAK, etc.)
    CRC         string     // Valor CRC32 para verificação de integridade
    Message     string     // Conteúdo da mensagem
    RawData     string     // Representação em string do pacote completo
}
```

### 2.2 Fila de Mensagens

```go
// Fila de mensagens para envio
type MessageQueue struct {
    messages []*message.QueuedMessage  // Slice de mensagens
    mutex    sync.RWMutex              // Mutex para acesso concorrente
    maxSize  int                       // Tamanho máximo da fila
}
```

### 2.3 Máquina da Rede

```go
// Status da máquina para monitoramento
type MachineStatus struct {
    MachineName      string
    HasToken         bool
    QueueSize        int
    LastActivity     time.Time
    TokensProcessed  int
    MessagesSent     int
    MessagesReceived int
    ErrorsDetected   int
    TokensGenerated  int
}

// Máquina da rede
type Machine struct {
    config           *config.Config
    conn             *net.UDPConn
    queue            *queue.MessageQueue
    hasToken         bool
    running          bool
    mutex            sync.RWMutex
    lastActivity     time.Time
    status           *MachineStatus
    tokenTimeout     *time.Timer
    waitingForData   bool
    currentDataMsg   *message.DataMessage
    errorProbability float64
}
```

## 3. Mecanismos de Sincronização

### 3.1 Mutex

O projeto utiliza `sync.RWMutex` para garantir acesso seguro a recursos compartilhados em ambiente concorrente:

- Na fila de mensagens (`MessageQueue`): Protege o acesso à lista de mensagens.
- Na máquina (`Machine`): Protege o acesso a variáveis como `hasToken`, `running`, `status`, etc.

### 3.2 WaitGroup

Utilizado no `main.go` para aguardar a finalização das goroutines:

```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    machine.Start()
}()
```

### 3.3 Timers

Utilizados para controlar o tempo de posse do token e para detectar tokens perdidos:

```go
m.tokenTimeout = time.AfterFunc(time.Duration(m.config.TokenTime)*time.Second, func() {
    m.processToken()
})
```

## 4. Threads (Goroutines)

O projeto utiliza múltiplas goroutines para operações concorrentes:

1. **Goroutine Principal**: Executa o método `Start()` da máquina, responsável por receber pacotes da rede.
2. **Goroutine de Interface**: Processa comandos do usuário via terminal.
3. **Goroutine de Timeout do Token**: Processa o token após o tempo configurado.
4. **Goroutine de Watchdog**: Monitora a circulação do token e gera um novo se necessário.

## 5. Implementação do CRC

O pacote `crc` implementa funções para cálculo e verificação de CRC32:

```go
// Calcula o CRC32 de uma string
func CalculateCRC32(data string) uint32 {
    return crc32.ChecksumIEEE([]byte(data))
}

// Verifica se o CRC32 calculado corresponde ao esperado
func VerifyCRC32(data string, expectedCRC uint32) bool {
    calculated := CalculateCRC32(data)
    return calculated == expectedCRC
}

// Cria a string para cálculo do CRC
func CreateDataForCRC(origin, destination, message string) string {
    return origin + ":" + destination + ":" + message
}
```

O CRC32 é utilizado para garantir a integridade das mensagens. Quando uma mensagem é recebida, o CRC é recalculado e comparado com o valor enviado. Se forem diferentes, a mensagem é considerada corrompida e um NAK é enviado.

## 6. Protocolo de Comunicação

### 6.1 Formato dos Pacotes

#### Token
```
1000
```

#### Dados
```
2000;origem:destino:controle:crc:mensagem
```

### 6.2 Campos de Controle

- `maquinanaoexiste`: Indica que a máquina de destino não existe
- `ACK`: Confirmação positiva de recebimento
- `NAK`: Confirmação negativa (erro detectado)

### 6.3 Fluxo de Comunicação

1. O token circula pela rede.
2. Quando uma máquina recebe o token, ela verifica se tem mensagens para enviar.
3. Se tiver, envia a primeira mensagem da fila e aguarda o retorno.
4. A mensagem circula pela rede até chegar ao destinatário.
5. O destinatário verifica a integridade da mensagem usando CRC32.
6. Se a mensagem estiver íntegra, o destinatário envia um ACK; caso contrário, envia um NAK.
7. A mensagem continua circulando até retornar à origem.
8. A origem processa o ACK/NAK e passa o token adiante.

## 7. Simulação de Erros

O sistema simula erros de transmissão com uma probabilidade configurável:

```go
func (dm *DataMessage) IntroduceError(probability float64) bool {
    if rand.Float64() < probability {
        // Corrompe o CRC da mensagem
        originalCRC := dm.CRC
        corruptedCRC := strconv.FormatUint(uint64(rand.Uint32()), 10)
        
        // Garante que o CRC corrompido seja diferente do original
        for corruptedCRC == originalCRC {
            corruptedCRC = strconv.FormatUint(uint64(rand.Uint32()), 10)
        }
        
        dm.CRC = corruptedCRC
        // Atualiza o pacote raw
        dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
            DataPacket, dm.Origin, dm.Destination, dm.Control, dm.CRC, dm.Message)
        
        return true
    }
    return false
}
```

## 8. Exemplos de Execução

### 8.1 Configuração

Exemplo de arquivo de configuração (`config_teste_local.txt`):
```
127.0.0.1:6001
Alice
3
true
```

Onde:
- `127.0.0.1:6001`: Endereço da próxima máquina
- `Alice`: Nome desta máquina
- `3`: Tempo de posse do token (segundos)
- `true`: Indica se esta máquina gera o token inicial

### 8.2 Inicialização

```
$ go run cmd/machine/main.go config_teste_local.txt

=== Iniciando Máquina da Rede em Anel ===
Máquina: Alice
Destino do token: 127.0.0.1:6001
Tempo do token: 3 segundos
Gera token inicial: true
=====================================

=== Interface de Comandos ===
Comandos disponíveis:
1. send <destino> <mensagem> - Enviar mensagem unicast
2. broadcast <mensagem> - Enviar mensagem broadcast
3. status - Ver status da máquina
4. queue - Ver fila de mensagens
5. token - Gerar novo token (se autorizado)
6. help - Mostrar comandos
7. logs - Ver últimas linhas do arquivo de log
8. quit - Sair
============================
```

### 8.3 Envio de Mensagem

```
> send Bob Olá, como vai?
Mensagem adicionada à fila para Bob: Olá, como vai?

> status
Status da Máquina:
  Nome: Alice
  Possui Token: false
  Mensagens na Fila: 1
  Última Atividade: 15:30:45
  Tokens Processados: 2
  Mensagens Enviadas: 0
  Mensagens Recebidas: 0
```

### 8.4 Recebimento de Mensagem

Quando a máquina recebe o token e tem mensagens na fila:

```
> logs
=== Últimas linhas do log ===
2023/06/10 15:30:50 [Alice] Token recebido
2023/06/10 15:30:50 [Alice] Enviando mensagem para Bob: Olá, como vai?
2023/06/10 15:30:53 [Alice] ACK recebido para mensagem para Bob
2023/06/10 15:30:53 [Alice] Token enviado para próxima máquina
================================
```

### 8.5 Detecção de Erro

Quando ocorre um erro na transmissão:

```
> logs
=== Últimas linhas do log ===
2023/06/10 15:35:20 [Alice] Token recebido
2023/06/10 15:35:20 [Alice] Enviando mensagem para Carol: Teste com erro
2023/06/10 15:35:20 [Alice] Erro introduzido na mensagem para Carol
2023/06/10 15:35:23 [Alice] NAK recebido para mensagem para Carol - será retransmitida
2023/06/10 15:35:23 [Alice] Token enviado para próxima máquina
=============================
```

## 9. Conclusão

A implementação apresentada demonstra um sistema de rede em anel com protocolo Token Ring, utilizando conceitos importantes de redes de computadores:

- **Controle de Acesso ao Meio**: Através do token que circula pela rede.
- **Detecção de Erros**: Utilizando CRC32 para verificar a integridade das mensagens.
- **Retransmissão**: Mensagens corrompidas são retransmitidas.
- **Comunicação Unicast e Broadcast**: Suporte para envio de mensagens para um destinatário específico ou para todos.
- **Recuperação de Falhas**: Mecanismo de watchdog para detectar e recuperar tokens perdidos.

O sistema é robusto e escalável, permitindo a adição de novas máquinas à rede com configurações simples. A interface de linha de comando facilita a interação do usuário com o sistema, permitindo o envio de mensagens e a visualização do estado da máquina.