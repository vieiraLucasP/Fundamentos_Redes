package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config representa a configuração da máquina
type Config struct {
	NextMachineAddr string // IP:porta da próxima máquina no anel
	MachineName     string // Apelido da máquina atual
	TokenTime       int    // Tempo que o token permanece na máquina (segundos)
	GeneratesToken  bool   // Se esta máquina gera o token inicial
	ListenPort      int    // Porta para escutar pacotes UDP
}

// LoadConfig carrega a configuração do arquivo especificado
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo de configuração: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Ler todas as linhas não vazias
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %v", err)
	}

	if len(lines) < 4 {
		return nil, fmt.Errorf("arquivo de configuração incompleto. Esperado 4 linhas, encontrado %d", len(lines))
	}

	// Parsear configuração
	cfg := &Config{}

	// Linha 1: IP:porta do destino
	cfg.NextMachineAddr = lines[0]

	// Linha 2: Apelido da máquina
	cfg.MachineName = lines[1]

	// Linha 3: Tempo do token
	tokenTime, err := strconv.Atoi(lines[2])
	if err != nil {
		return nil, fmt.Errorf("tempo do token inválido: %v", err)
	}
	cfg.TokenTime = tokenTime

	// Linha 4: Gera token inicial
	generatesToken, err := strconv.ParseBool(lines[3])
	if err != nil {
		return nil, fmt.Errorf("valor de geração de token inválido: %v", err)
	}
	cfg.GeneratesToken = generatesToken

	// Determinar porta de escuta baseada no nome da máquina
	switch cfg.MachineName {
	case "Alice":
		cfg.ListenPort = 6000
	case "Bob":
		cfg.ListenPort = 6001
	case "Carol":
		cfg.ListenPort = 6002
	default:
		// Para outras máquinas, extrair porta do endereço de destino como fallback
		parts := strings.Split(cfg.NextMachineAddr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("formato de endereço inválido: %s", cfg.NextMachineAddr)
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("porta inválida: %v", err)
		}
		cfg.ListenPort = port - 1 // Usar porta anterior como convenção
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
	return fmt.Sprintf("Config{NextMachine: %s, Name: %s, TokenTime: %d, GeneratesToken: %t, ListenPort: %d}",
		c.NextMachineAddr, c.MachineName, c.TokenTime, c.GeneratesToken, c.ListenPort)
}
