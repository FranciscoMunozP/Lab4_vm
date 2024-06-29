# Nombres de los binarios a generar
BROKER_BIN=bin/Broker

# Directorios fuente
BROKER_SRC_DIR=Broker

# Variables de configuración de Go
GO=go
GO_FLAGS=

# Tarea por defecto
all: build

# Construir los binarios
build: build-Broker

build-Broker:
	@echo "Compilando Broker..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(BROKER_BIN) $(BROKER_SRC_DIR)/Broker.go

# Ejecutar los binarios en nuevas ventanas de terminal
run: run-Broker

run-Broker:
	@echo "Ejecutando Broker en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(BROKER_BIN); exec bash"

# Limpiar archivos generados
clean:
	@echo "Limpiando binarios generados..."
	rm -rf bin

# Mostrar ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make build        - Construir todos los binarios"
	@echo "  make build-Broker - Construir el binario de Broker"
	@echo "  make run          - Ejecutar todos los binarios en nuevas ventanas de terminal"
	@echo "  make run-Broker   - Ejecutar Broker en una nueva ventana de terminal"
	@echo "  make clean        - Limpiar binarios generados"
	@echo "  make help         - Mostrar este mensaje de ayuda"

.PHONY: all build build-Broker build-server2 run run-Broker run-server2 clean help
