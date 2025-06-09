# Makefile para Simulação de Rede em Anel

.PHONY: build run clean fmt vet help alice bob carol

# Variáveis
BINARY_NAME=machine
BUILD_DIR=bin
CMD_DIR=cmd/machine

# Comandos principais
help: ## Mostrar ajuda
	@echo "Comandos disponíveis:"
	@echo "  build   - Compilar a aplicação"
	@echo "  run     - Executar com configuração padrão (Alice)"
	@echo "  clean   - Limpar arquivos compilados"
	@echo "  fmt     - Formatar código"
	@echo "  vet     - Verificar código"
	@echo "  alice   - Executar máquina Alice (gera token)"
	@echo "  bob     - Executar máquina Bob"
	@echo "  carol   - Executar máquina Carol"

build: ## Compilar a aplicação
	@echo "Compilando aplicação..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

run: build ## Executar com configuração padrão
	@echo "Executando máquina Alice..."
	./$(BUILD_DIR)/$(BINARY_NAME) alice.txt

clean: ## Limpar arquivos compilados
	@echo "Limpando arquivos..."
	rm -rf $(BUILD_DIR)
	go clean

fmt: ## Formatar código
	@echo "Formatando código..."
	go fmt ./...

vet: ## Verificar código
	@echo "Verificando código..."
	go vet ./...

# Comandos para executar diferentes máquinas
alice: build ## Executar máquina Alice (gera token inicial)
	@echo "Iniciando máquina Alice (porta 6000)..."
	./$(BUILD_DIR)/$(BINARY_NAME) alice.txt

bob: build ## Executar máquina Bob
	@echo "Iniciando máquina Bob (porta 6001)..."
	./$(BUILD_DIR)/$(BINARY_NAME) bob.txt

carol: build ## Executar máquina Carol
	@echo "Iniciando máquina Carol (porta 6002)..."
	./$(BUILD_DIR)/$(BINARY_NAME) carol.txt

# Comandos de desenvolvimento
deps: ## Baixar dependências
	@echo "Baixando dependências..."
	go mod tidy
	go mod download