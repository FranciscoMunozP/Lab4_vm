# Nombres de los binarios a generar
CAPITAN_BIN=bin/Capitan
SERVER2_BIN=bin/server2

# Directorios fuente
CAPITAN_SRC_DIR=capitan
SERVER2_SRC_DIR=server2

# Variables de configuración de Go
GO=go
GO_FLAGS=

# Tarea por defecto
all: build

# Construir los binarios
build: build-capitan build-server2

build-capitan:
	@echo "Compilando capitan..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(CAPITAN_BIN) $(CAPITAN_SRC_DIR)/capitan.go

build-server2:
	@echo "Compilando server2..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(SERVER2_BIN) $(SERVER2_SRC_DIR)/server2.go

# Ejecutar los binarios en nuevas ventanas de terminal
run: run-capitan run-server2

run-capitan:
	@echo "Ejecutando capitan en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(CAPITAN_BIN); exec bash"

run-server2:
	@echo "Ejecutando server2 en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(SERVER2_BIN); exec bash"

# Limpiar archivos generados
clean:
	@echo "Limpiando binarios generados..."
	rm -rf bin

# Mostrar ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make build        - Construir todos los binarios"
	@echo "  make build-capitan - Construir el binario de capitan"
	@echo "  make build-server2- Construir el binario de server2"
	@echo "  make run          - Ejecutar todos los binarios en nuevas ventanas de terminal"
	@echo "  make run-capitan   - Ejecutar capitan en una nueva ventana de terminal"
	@echo "  make run-server2  - Ejecutar server2 en una nueva ventana de terminal"
	@echo "  make clean        - Limpiar binarios generados"
	@echo "  make help         - Mostrar este mensaje de ayuda"

.PHONY: all build build-capitan build-server2 run run-capitan run-server2 clean help
