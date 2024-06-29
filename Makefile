# Nombres de los binarios a generar
JETH_BIN=bin/Jeth
SERVER1_BIN=bin/server1

# Directorios fuente
JETH_SRC_DIR=Jeth
SERVER1_SRC_DIR=server1

# Variables de configuración de Go
GO=go
GO_FLAGS=

# Tarea por defecto
all: build

# Construir los binarios
build: build-Jeth build-server1

build-Jeth:
	@echo "Compilando Jeth..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(JETH_BIN) $(JETH_SRC_DIR)/Jeth.go

build-server1:
	@echo "Compilando server1..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(SERVER1_BIN) $(SERVER1_SRC_DIR)/server1.go

# Ejecutar los binarios en nuevas ventanas de terminal
run: run-Jeth run-server1

run-Jeth:
	@echo "Ejecutando Jeth en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(JETH_BIN); exec bash"

run-server1:
	@echo "Ejecutando server1 en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(SERVER1_BIN); exec bash"

# Limpiar archivos generados
clean:
	@echo "Limpiando binarios generados..."
	rm -rf bin

# Mostrar ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make build        - Construir todos los binarios"
	@echo "  make build-Jeth - Construir el binario de Jeth"
	@echo "  make build-server1- Construir el binario de server1"
	@echo "  make run          - Ejecutar todos los binarios en nuevas ventanas de terminal"
	@echo "  make run-Jeth   - Ejecutar Jeth en una nueva ventana de terminal"
	@echo "  make run-server1  - Ejecutar server1 en una nueva ventana de terminal"
	@echo "  make clean        - Limpiar binarios generados"
	@echo "  make help         - Mostrar este mensaje de ayuda"

.PHONY: all build build-Jeth build-server1 run run-Jeth run-server1 clean help
