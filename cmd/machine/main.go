package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"ring-network/pkg/config"
	"ring-network/pkg/network"
)

// main é o ponto de entrada da aplicação
// Inicializa a máquina da rede em anel e a interface de comandos
func main() {
	// Verifica se foi fornecido o arquivo de configuração
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <arquivo_de_configuracao>")
		fmt.Println("Exemplo: go run main.go config.txt")
		os.Exit(1)
	}

	configFile := os.Args[1]

	// Carrega a configuração do arquivo
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Configura o sistema de log
	if err := cfg.SetupLogger(); err != nil {
		fmt.Printf("Aviso: Não foi possível configurar o arquivo de log: %v\n", err)
		fmt.Println("Os logs serão exibidos apenas no terminal.")
	} else {
		fmt.Printf("Logs sendo gravados em: %s\n", cfg.LogFile)
		fmt.Println("O terminal agora está limpo para comandos.")
	}

	// Exibe informações de inicialização
	fmt.Printf("=== Iniciando Máquina da Rede em Anel ===\n")
	fmt.Printf("Máquina: %s\n", cfg.MachineName)
	fmt.Printf("Destino do token: %s\n", cfg.NextMachineAddr)
	fmt.Printf("Tempo do token: %d segundos\n", cfg.TokenTime)
	fmt.Printf("Gera token inicial: %t\n", cfg.GeneratesToken)
	fmt.Println("=====================================")

	// Cria a máquina com a configuração carregada
	machine, err := network.NewMachine(cfg)
	if err != nil {
		log.Fatalf("Erro ao criar máquina: %v", err)
	}

	// Inicia a máquina em uma goroutine separada
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		machine.Start()
	}()

	// Inicia a interface de comandos em outra goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("\n=== Interface de Comandos ===")
		fmt.Println("Comandos disponíveis:")
		fmt.Println("1. send <destino> <mensagem> - Enviar mensagem unicast")
		fmt.Println("2. broadcast <mensagem> - Enviar mensagem broadcast")
		fmt.Println("3. status - Ver status da máquina")
		fmt.Println("4. queue - Ver fila de mensagens")
		fmt.Println("5. token - Gerar novo token (se autorizado)")
		fmt.Println("6. help - Mostrar comandos")
		fmt.Println("7. logs - Ver últimas linhas do arquivo de log")
		fmt.Println("8. quit - Sair")
		fmt.Println("============================")

		// Loop principal da interface de comandos
		for {
			fmt.Print("\n> ")
			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Divide a entrada em partes para processamento
			parts := strings.SplitN(input, " ", 3)
			command := strings.ToLower(parts[0])

			// Processa o comando
			switch command {
			case "send":
				// Envia mensagem unicast
				if len(parts) < 3 {
					fmt.Println("Uso: send <destino> <mensagem>")
					continue
				}
				destination := parts[1]
				message := parts[2]
				err := machine.QueueMessage(destination, message)
				if err != nil {
					fmt.Printf("Erro ao enviar mensagem: %v\n", err)
				} else {
					fmt.Printf("Mensagem adicionada à fila para %s: %s\n", destination, message)
				}

			case "broadcast":
				// Envia mensagem broadcast
				if len(parts) < 2 {
					fmt.Println("Uso: broadcast <mensagem>")
					continue
				}
				message := strings.Join(parts[1:], " ")
				err := machine.QueueMessage("TODOS", message)
				if err != nil {
					fmt.Printf("Erro ao enviar broadcast: %v\n", err)
				} else {
					fmt.Printf("Mensagem broadcast adicionada à fila: %s\n", message)
				}

			case "status":
				// Exibe o status da máquina
				status := machine.GetStatus()
				fmt.Printf("Status da Máquina:\n")
				fmt.Printf("  Nome: %s\n", status.MachineName)
				fmt.Printf("  Possui Token: %t\n", status.HasToken)
				fmt.Printf("  Mensagens na Fila: %d\n", status.QueueSize)
				fmt.Printf("  Última Atividade: %s\n", status.LastActivity.Format("15:04:05"))
				fmt.Printf("  Tokens Processados: %d\n", status.TokensProcessed)
				fmt.Printf("  Mensagens Enviadas: %d\n", status.MessagesSent)
				fmt.Printf("  Mensagens Recebidas: %d\n", status.MessagesReceived)

			case "queue":
				// Exibe a fila de mensagens
				queue := machine.GetMessageQueue()
				if len(queue) == 0 {
					fmt.Println("Fila de mensagens vazia")
				} else {
					fmt.Printf("Fila de mensagens (%d/%d):\n", len(queue), 10)
					for i, msg := range queue {
						fmt.Printf("  %d. Para: %s | Mensagem: %s\n", i+1, msg.Destination, msg.Content)
					}
				}

			case "token":
				// Gera um novo token
				err := machine.GenerateToken()
				if err != nil {
					fmt.Printf("Erro ao gerar token: %v\n", err)
				} else {
					fmt.Println("Novo token gerado e enviado")
				}

			case "help":
				// Exibe ajuda
				fmt.Println("\nComandos disponíveis:")
				fmt.Println("1. send <destino> <mensagem> - Enviar mensagem unicast")
				fmt.Println("2. broadcast <mensagem> - Enviar mensagem broadcast")
				fmt.Println("3. status - Ver status da máquina")
				fmt.Println("4. queue - Ver fila de mensagens")
				fmt.Println("5. token - Gerar novo token (se autorizado)")
				fmt.Println("6. help - Mostrar comandos")
				fmt.Println("7. logs - Ver últimas linhas do arquivo de log")
				fmt.Println("8. quit - Sair")

			case "logs":
				// Exibe as últimas linhas do arquivo de log
				if cfg.LogFile == "" {
					fmt.Println("Logs não estão sendo gravados em arquivo.")
					continue
				}

				file, err := os.Open(cfg.LogFile)
				if err != nil {
					fmt.Printf("Erro ao abrir arquivo de log: %v\n", err)
					continue
				}
				defer file.Close()

				// Lê as últimas 20 linhas do arquivo
				scanner := bufio.NewScanner(file)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
					if len(lines) > 20 {
						lines = lines[1:]
					}
				}

				if err := scanner.Err(); err != nil {
					fmt.Printf("Erro ao ler arquivo de log: %v\n", err)
					continue
				}

				// Exibe as linhas
				fmt.Println("\n=== Últimas linhas do log ===")
				if len(lines) == 0 {
					fmt.Println("Nenhum log encontrado.")
				} else {
					for _, line := range lines {
						fmt.Println(line)
					}
				}
				fmt.Println("=============================")

			case "quit", "exit":
				// Encerra a máquina
				fmt.Println("Encerrando máquina...")
				machine.Stop()
				os.Exit(0)

			default:
				fmt.Printf("Comando desconhecido: %s. Digite 'help' para ver os comandos disponíveis.\n", command)
			}
		}
	}()

	// Aguarda a finalização da goroutine principal
	wg.Wait()
}
