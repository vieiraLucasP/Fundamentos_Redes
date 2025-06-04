# 🌐 GUIA DE CONFIGURAÇÃO PARA REDE FÍSICA
# ========================================
# Cada aluno configura APENAS sua própria máquina

## 📋 PASSO A PASSO PARA CONFIGURAÇÃO

### 1. DESCOBRIR SEU IP
Execute em sua máquina:
```bash
# macOS/Linux
ifconfig | grep "inet " | grep -v 127.0.0.1

# Windows
ipconfig | findstr "IPv4"
```

### 2. DEFINIR ORDEM DO ANEL
Combinem entre vocês a ordem das máquinas:
- Máquina 1: Alice (gera token) → IP: 192.168.1.101
- Máquina 2: Bob (intermediária) → IP: 192.168.1.102  
- Máquina 3: Carol (fecha anel) → IP: 192.168.1.103

### 3. CRIAR SEU ARQUIVO DE CONFIGURAÇÃO

Se você for a ALICE (primeira máquina):
```
192.168.1.102:6001
Alice
3
true
```

Se você for o BOB (segunda máquina):
```
192.168.1.103:6002
Bob
3
false
```

Se você for a CAROL (terceira máquina):
```
192.168.1.101:6000
Carol
3
false
```

### 4. FORMATO EXPLICADO
```
<IP_DA_PRÓXIMA_MÁQUINA>:porta
<SEU_NOME_DE_MÁQUINA>
<TEMPO_DO_TOKEN_EM_SEGUNDOS>
<GERA_TOKEN_INICIAL: true apenas para Alice>
```

### 5. PORTAS UTILIZADAS
- Alice escuta na porta: 6000
- Bob escuta na porta: 6001
- Carol escuta na porta: 6002

### 6. TESTAR CONECTIVIDADE
Antes de executar, testem a conectividade:
```bash
# Cada um pinga os outros
ping 192.168.1.101  # Alice
ping 192.168.1.102  # Bob
ping 192.168.1.103  # Carol
```

### 7. EXECUTAR
Cada um executa em sua máquina:
```bash
# Use o arquivo de exemplo como base e crie seu próprio
cp config_exemplo.txt minha_config.txt
# Edite com seus valores e execute:
./bin/machine minha_config.txt
```

## 📁 ARQUIVO DE EXEMPLO
Veja `config_exemplo.txt` como referência para criar sua configuração.

## ⚠️  IMPORTANTE
- Apenas UMA pessoa deve ser Alice (true no último campo)
- As outras duas devem ter false
- Todos devem estar na mesma rede Wi-Fi
- IPs devem ser descobertos na hora (podem mudar)

## 🔧 TROUBLESHOOTING
- Firewall: Liberem as portas 6000-6002 UDP
- Antivírus: Podem bloquear conexões UDP
- Rede corporativa: Pode bloquear comunicação entre máquinas
