# Makefile para Simulação de Rede em Anel

.PHONY: build run clean test fmt vet help alice bob carol demo

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
	@echo "  test    - Executar testes"
	@echo "  fmt     - Formatar código"
	@echo "  vet     - Verificar código"
	@echo "  alice   - Executar máquina Alice (gera token)"
	@echo "  bob     - Executar máquina Bob"
	@echo "  carol   - Executar máquina Carol"
	@echo "  demo    - Executar demonstração (3 terminais)"

build: ## Compilar a aplicação
	@echo "Compilando aplicação..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

run: build ## Executar com configuração padrão
	@echo "Executando máquina Alice..."
	./$(BUILD_DIR)/$(BINARY_NAME) config_alice.txt

clean: ## Limpar arquivos compilados
	@echo "Limpando arquivos..."
	rm -rf $(BUILD_DIR)
	go clean

test: ## Executar testes
	@echo "Executando testes..."
	go test -v ./...

fmt: ## Formatar código
	@echo "Formatando código..."
	go fmt ./...

vet: ## Verificar código
	@echo "Verificando código..."
	go vet ./...

# Comandos para executar diferentes máquinas
alice: ## Executar máquina Alice (gera token inicial)
	@echo "Iniciando máquina Alice (porta 6000)..."
	go run $(CMD_DIR)/main.go config_alice.txt

bob: ## Executar máquina Bob
	@echo "Iniciando máquina Bob (porta 6001)..."
	go run $(CMD_DIR)/main.go config_bob.txt

carol: ## Executar máquina Carol
	@echo "Iniciando máquina Carol (porta 6002)..."
	go run $(CMD_DIR)/main.go config_carol.txt

# Demonstração automatizada
demo: ## Executar demonstração em 3 terminais
	@echo "Iniciando demonstração..."
	@echo "Execute os seguintes comandos em terminais separados:"
	@echo ""
	@echo "Terminal 1 (Alice - gera token):"
	@echo "make alice"
	@echo ""
	@echo "Terminal 2 (Bob):"
	@echo "make bob"
	@echo ""
	@echo "Terminal 3 (Carol):"
	@echo "make carol"
	@echo ""
	@echo "Aguarde alguns segundos entre cada inicialização!"

# Comandos de desenvolvimento
deps: ## Baixar dependências
	@echo "Baixando dependências..."
	go mod tidy
	go mod download

check: fmt vet test ## Executar todas as verificações

install: build ## Instalar no sistema
	@echo "Instalando aplicação..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
