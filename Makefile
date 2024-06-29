# Nombres de los binarios a generar
MALKOR_BIN=bin/Malkor
SERVER0_BIN=bin/server0

# Directorios fuente
MALKOR_SRC_DIR=Malkor
SERVER0_SRC_DIR=server0

# Variables de configuración de Go
GO=go
GO_FLAGS=

# Tarea por defecto
all: build

# Construir los binarios
build: build-malkor build-server0

build-malkor:
	@echo "Compilando Malkor..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(MALKOR_BIN) $(MALKOR_SRC_DIR)/Malkor.go

build-server0:
	@echo "Compilando server0..."
	mkdir -p bin
	$(GO) build $(GO_FLAGS) -o $(SERVER0_BIN) $(SERVER0_SRC_DIR)/server0.go

# Ejecutar los binarios en nuevas ventanas de terminal
run: run-malkor run-server0

run-malkor:
	@echo "Ejecutando Malkor en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(MALKOR_BIN); exec bash"

run-server0:
	@echo "Ejecutando server0 en una nueva ventana de terminal..."
	gnome-terminal -- bash -c "$(PWD)/$(SERVER0_BIN); exec bash"

# Limpiar archivos generados
clean:
	@echo "Limpiando binarios generados..."
	rm -rf bin

# Mostrar ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make build        - Construir todos los binarios"
	@echo "  make build-malkor - Construir el binario de Malkor"
	@echo "  make build-server0- Construir el binario de server0"
	@echo "  make run          - Ejecutar todos los binarios en nuevas ventanas de terminal"
	@echo "  make run-malkor   - Ejecutar Malkor en una nueva ventana de terminal"
	@echo "  make run-server0  - Ejecutar server0 en una nueva ventana de terminal"
	@echo "  make clean        - Limpiar binarios generados"
	@echo "  make help         - Mostrar este mensaje de ayuda"

.PHONY: all build build-malkor build-server0 run run-malkor run-server0 clean help
