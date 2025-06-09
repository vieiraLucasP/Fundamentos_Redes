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

type Config struct {
	NextMachineAddr string
	MachineName     string
	TokenTime       int
	GeneratesToken  bool
	ListenPort      int
	LogFile         string
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo de configuração: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

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

	cfg := &Config{}

	cfg.NextMachineAddr = lines[0]

	cfg.MachineName = lines[1]

	tokenTime, err := strconv.Atoi(lines[2])
	if err != nil {
		return nil, fmt.Errorf("tempo do token inválido: %v", err)
	}
	cfg.TokenTime = tokenTime

	generatesToken, err := strconv.ParseBool(lines[3])
	if err != nil {
		return nil, fmt.Errorf("valor de geração de token inválido: %v", err)
	}
	cfg.GeneratesToken = generatesToken

	cfg.LogFile = fmt.Sprintf("%s_log.txt", strings.ToLower(cfg.MachineName))

	switch cfg.MachineName {
	case "Alice":
		cfg.ListenPort = 6000
	case "Bob":
		cfg.ListenPort = 6001
	case "Carol":
		cfg.ListenPort = 6002
	default:
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

func (c *Config) String() string {
	return fmt.Sprintf("Config{NextMachine: %s, Name: %s, TokenTime: %d, GeneratesToken: %t, ListenPort: %d, LogFile: %s}",
		c.NextMachineAddr, c.MachineName, c.TokenTime, c.GeneratesToken, c.ListenPort, c.LogFile)
}
func (c *Config) SetupLogger() error {
	logDir := filepath.Dir(c.LogFile)
	if logDir != "." && logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório de logs: %v", err)
		}
	}

	logFile, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de log: %v", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("=== Iniciando logs para máquina %s ===", c.MachineName)

	return nil
}
