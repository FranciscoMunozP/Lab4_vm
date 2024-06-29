package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/FranciscoMunozP/Lab5_proto"
)

type server struct {
	pb.UnimplementedFulcrumServer
	mu        sync.Mutex
	serverID  int
	reloj     [3]string
	serverlog []string
}

func (s *server) propagarCambios() error {
	Add := []string{}
	var flag int
	for _, lines := range s.serverlog {
		fmt.Printf("Log: %s\n", lines)
		parts := strings.Fields(lines)
		fmt.Printf("Func: %s\n", parts[0])
		switch parts[0] {
		case "AgregarBase":
			if len(Add) == 0 {
				flag = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
			} else {
				for _, v := range Add {
					if parts[1] == v {
						flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
						break
					} else {
						flag = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
					}
				}
			}
			Add = append(Add, parts[1])
			sectorFile := fmt.Sprintf("./Sector%s.txt", parts[1])
			file, err := os.OpenFile(sectorFile, flag, 0644)
			if err != nil {
				return err
			}
			cantidad := "0"
			if len(parts) == 4 {
				cantidad = parts[3]
			}

			baseEntry := fmt.Sprintf("%s %s %s\n", parts[1], parts[2], cantidad)
			_, err = file.WriteString(baseEntry)
			if err != nil {
				return err
			}

			file.Close()
		case "RenombrarBase":
			sectorFile := fmt.Sprintf("./Sector%s.txt", parts[1])
			input, err := os.ReadFile(sectorFile)
			if err != nil {
				return err
			}

			lines := strings.Split(string(input), "\n")
			for i, line := range lines {
				aux := strings.Fields(line)
				if len(aux) >= 3 && aux[1] == parts[2] {
					lines[i] = fmt.Sprintf("%s %s %s", parts[1], parts[2], parts[3])
				}
			}

			output := strings.Join(lines, "\n")
			err = os.WriteFile(sectorFile, []byte(output), 0644)
			if err != nil {
				return err
			}
		case "ActualizarValor":
			sectorFile := fmt.Sprintf("./Sector%s.txt", parts[1])
			input, err := os.ReadFile(sectorFile)
			if err != nil {
				return err
			}

			lines := strings.Split(string(input), "\n")
			for i, line := range lines {
				aux := strings.Fields(line)
				if len(aux) >= 3 && aux[1] == parts[2] {
					lines[i] = fmt.Sprintf("%s %s %s", parts[1], parts[2], parts[3])
				}
			}

			output := strings.Join(lines, "\n")
			err = os.WriteFile(sectorFile, []byte(output), 0644)
			if err != nil {
				return err
			}
		case "BorrarBase":
			sectorFile := fmt.Sprintf("./Sector%s.txt", parts[1])
			input, err := os.ReadFile(sectorFile)
			if err != nil {
				return err
			}

			lines := strings.Split(string(input), "\n")
			for i, line := range lines {
				aux := strings.Fields(line)
				if len(aux) >= 3 && aux[1] == parts[2] {
					lines = append(lines[:i], lines[i+1:]...)
					break
				}
			}

			output := strings.Join(lines, "\n")
			err = os.WriteFile(sectorFile, []byte(output), 0644)
			if err != nil {
				return err
			}
		}
	}
	s.serverlog = []string{}
	return nil
}

func (s *server) Sincronizar(ctx context.Context, in *pb.SyncRequest) (*pb.RespuestaSync, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, v := range in.RelojVector {
		ri, _ := strconv.Atoi(v)
		si, _ := strconv.Atoi(s.reloj[i])
		if ri > si {
			s.reloj[i] = v
		}
	}
	var combinados []string

	combinados = append(combinados, s.serverlog...)
	combinados = append(combinados, in.Logs...)
	s.serverlog = combinados

	return &pb.RespuestaSync{
		RelojVector: s.reloj[:],
		Logs:        s.serverlog,
	}, nil
}

func (s *server) escribirLog(comando string, parametro string) {
	aux := comando + parametro
	s.serverlog = append(s.serverlog, aux)
	addtime, _ := strconv.Atoi(s.reloj[s.serverID])
	addtime++
	s.reloj[s.serverID] = strconv.Itoa(addtime)
}

func (s *server) AgregarBase(ctx context.Context, in *pb.DatosSector) (*pb.RespuestaReloj, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Mensaje: AgregarBase %s %s %s\n", in.NombreSector, in.NombreBase, in.CantidadEnemigos)

	sectorFile := fmt.Sprintf("./Sector%s.txt", in.NombreSector)
	file, err := os.OpenFile(sectorFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	cantidad := "0"
	if in.CantidadEnemigos != "" {
		cantidad = in.CantidadEnemigos
	}

	baseEntry := fmt.Sprintf("%s %s %s\n", in.NombreSector, in.NombreBase, cantidad)
	_, err = file.WriteString(baseEntry)
	if err != nil {
		return nil, err
	}

	file.Close()
	s.escribirLog("AgregarBase", baseEntry)
	return &pb.RespuestaReloj{
		RelojVector: s.reloj[:],
	}, nil
}

func (s *server) RenombrarBase(ctx context.Context, in *pb.DatosSectorRenombrar) (*pb.RespuestaReloj, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Mensaje: RenombrarBase %s %s %s\n", in.NombreSector, in.NombreBaseAntigua, in.NombreBaseNueva)

	sectorFile := fmt.Sprintf("./Sector%s.txt", in.NombreSector)
	input, err := os.ReadFile(sectorFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[1] == in.NombreBaseAntigua {
			lines[i] = fmt.Sprintf("%s %s %s", in.NombreSector, in.NombreBaseNueva, parts[2])
		}
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(sectorFile, []byte(output), 0644)
	if err != nil {
		return nil, err
	}

	logEntry := fmt.Sprintf("%s %s %s", in.NombreSector, in.NombreBaseAntigua, in.NombreBaseNueva)
	s.escribirLog("RenombrarBase", logEntry)

	return &pb.RespuestaReloj{
		RelojVector: s.reloj[:],
	}, nil
}

func (s *server) ActualizarValor(ctx context.Context, in *pb.DatosSectorActualizar) (*pb.RespuestaReloj, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Mensaje: ActualizarValor %s %s %s\n", in.NombreSector, in.NombreBase, in.CantidadEnemigos)

	sectorFile := fmt.Sprintf("./Sector%s.txt", in.NombreSector)
	input, err := os.ReadFile(sectorFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[1] == in.NombreBase {
			lines[i] = fmt.Sprintf("%s %s %s", in.NombreSector, in.NombreBase, in.CantidadEnemigos)
		}
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(sectorFile, []byte(output), 0644)
	if err != nil {
		return nil, err
	}

	logEntry := fmt.Sprintf("%s %s %s", in.NombreSector, in.NombreBase, in.CantidadEnemigos)
	s.escribirLog("ActualizarValor", logEntry)
	return &pb.RespuestaReloj{
		RelojVector: s.reloj[:],
	}, nil
}

func (s *server) BorrarBase(ctx context.Context, in *pb.DatosSectorConsulta) (*pb.RespuestaReloj, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Mensaje: BorrarBase %s %s\n", in.NombreSector, in.NombreBase)

	sectorFile := fmt.Sprintf("./Sector%s.txt", in.NombreSector)
	input, err := os.ReadFile(sectorFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[1] == in.NombreBase {
			lines = append(lines[:i], lines[i+1:]...)
			break
		}
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(sectorFile, []byte(output), 0644)
	if err != nil {
		return nil, err
	}

	logEntry := fmt.Sprintf("%s %s", in.NombreSector, in.NombreBase)
	s.escribirLog("BorrarBase", logEntry)
	return &pb.RespuestaReloj{
		RelojVector: s.reloj[:],
	}, nil
}

func (s *server) GetEnemigos(ctx context.Context, in *pb.DatosSectorConsulta) (*pb.CantidadEnemigosResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Mensaje: GetEnemigos %s %s\n", in.NombreSector, in.NombreBase)

	sectorFile := fmt.Sprintf("./Sector%s.txt", in.NombreSector)
	input, err := os.Open(sectorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &pb.CantidadEnemigosResponse{CantidadEnemigos: "0"}, nil
		}
		return nil, err
	}
	defer input.Close()

	content, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[1] == in.NombreBase {
			return &pb.CantidadEnemigosResponse{
				CantidadEnemigos: parts[2],
				RelojVector:      s.reloj[:],
			}, nil
		}
	}
	return &pb.CantidadEnemigosResponse{
		CantidadEnemigos: "0",
	}, nil
}

func (s *server) Propagar(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	s.propagarCambios()
	return &emptypb.Empty{}, nil
}

func main() {
	serverID := 0
	reloj := [3]string{"2", "0", "0"}
	serverlog := []string{"AgregarBase Alpha Base1 90", "AgregarBase Alpha Base2 90"}

	s := &server{
		serverID:  serverID,
		reloj:     reloj,
		serverlog: serverlog,
	}

	lisClient, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051+serverID))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	lisReplica, err := net.Listen("tcp", fmt.Sprintf(":%d", 50061))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServerClient := grpc.NewServer()
	pb.RegisterFulcrumServer(grpcServerClient, s)

	grpcServerReplica := grpc.NewServer()
	pb.RegisterFulcrumServer(grpcServerReplica, s)

	go func() {
		log.Printf("Client server %d listening at %v", serverID, lisClient.Addr())
		if err := grpcServerClient.Serve(lisClient); err != nil {
			log.Fatalf("Failed to serve client connections: %v", err)
		}
	}()

	go func() {
		log.Printf("Replica server %d listening at %v", serverID, lisReplica.Addr())
		if err := grpcServerReplica.Serve(lisReplica); err != nil {
			log.Fatalf("Failed to serve replica connections: %v", err)
		}
	}()

	select {}
}
