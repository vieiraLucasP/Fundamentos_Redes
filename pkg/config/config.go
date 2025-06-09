package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config armazena as configurações de uma máquina na rede em anel
type Config struct {
	NextMachineAddr string // Endereço da próxima máquina na rede (IP:porta)
	MachineName     string // Nome da máquina atual
	TokenTime       int    // Tempo em segundos que a máquina pode reter o token
	GeneratesToken  bool   // Indica se esta máquina gera o token inicial
	ListenPort      int    // Porta em que a máquina escuta por conexões
	LogFile         string // Caminho do arquivo de log
}

// LoadConfig carrega as configurações a partir de um arquivo
// O arquivo deve conter pelo menos 4 linhas não comentadas:
// 1. Endereço da próxima máquina (IP:porta)
// 2. Nome desta máquina
// 3. Tempo do token em segundos
// 4. Flag indicando se gera token inicial (true/false)
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo de configuração: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Lê as linhas não vazias e não comentadas
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %v", err)
	}

	// Verifica se há linhas suficientes
	if len(lines) < 4 {
		return nil, fmt.Errorf("arquivo de configuração incompleto. Esperado 4 linhas, encontrado %d", len(lines))
	}

	cfg := &Config{}

	// Endereço da próxima máquina
	cfg.NextMachineAddr = lines[0]

	// Nome desta máquina
	cfg.MachineName = lines[1]

	// Tempo do token em segundos
	tokenTime, err := strconv.Atoi(lines[2])
	if err != nil {
		return nil, fmt.Errorf("tempo do token inválido: %v", err)
	}
	cfg.TokenTime = tokenTime

	// Flag para geração do token inicial
	generatesToken, err := strconv.ParseBool(lines[3])
	if err != nil {
		return nil, fmt.Errorf("valor de geração de token inválido: %v", err)
	}
	cfg.GeneratesToken = generatesToken

	// Define o arquivo de log baseado no nome da máquina
	cfg.LogFile = fmt.Sprintf("%s_log.txt", strings.ToLower(cfg.MachineName))

	// Define a porta de escuta com base no nome da máquina ou no endereço da próxima
	switch cfg.MachineName {
	case "Alice":
		cfg.ListenPort = 6000
	case "Bob":
		cfg.ListenPort = 6001
	case "Carol":
		cfg.ListenPort = 6002
	default:
		// Para outras máquinas, usa a porta da próxima máquina - 1
		parts := strings.Split(cfg.NextMachineAddr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("formato de endereço inválido: %s", cfg.NextMachineAddr)
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("porta inválida: %v", err)
		}
		cfg.ListenPort = port - 1
	}

	return cfg, nil
}

// Validate verifica se a configuração é válida
func (c *Config) Validate() error {
	if c.NextMachineAddr == "" {
		return fmt.Errorf("endereço da próxima máquina não pode estar vazio")
	}

	if c.MachineName == "" {
		return fmt.Errorf("nome da máquina não pode estar vazio")
	}

	if c.TokenTime <= 0 {
		return fmt.Errorf("tempo do token deve ser maior que zero")
	}

	if c.ListenPort <= 0 || c.ListenPort > 65535 {
		return fmt.Errorf("porta deve estar entre 1 e 65535")
	}

	return nil
}

// String retorna uma representação em string da configuração
func (c *Config) String() string {
	return fmt.Sprintf("Config{NextMachine: %s, Name: %s, TokenTime: %d, GeneratesToken: %t, ListenPort: %d, LogFile: %s}",
		c.NextMachineAddr, c.MachineName, c.TokenTime, c.GeneratesToken, c.ListenPort, c.LogFile)
}

// SetupLogger configura o sistema de log para gravar em arquivo
func (c *Config) SetupLogger() error {
	// Cria o diretório de logs se necessário
	logDir := filepath.Dir(c.LogFile)
	if logDir != "." && logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório de logs: %v", err)
		}
	}

	// Abre o arquivo de log (cria se não existir, trunca se existir)
	logFile, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de log: %v", err)
	}

	// Configura o logger padrão para escrever no arquivo
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("=== Iniciando logs para máquina %s ===", c.MachineName)

	return nil
}
