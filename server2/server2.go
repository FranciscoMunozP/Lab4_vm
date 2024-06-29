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
	"time"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/FranciscoMunozP/Lab5_proto"
)

type server struct {
	pb.UnimplementedFulcrumServer
	mu        sync.Mutex
	serverID  int
	reloj     [3]string
	serverlog []string
}

func (s *server) escribirLog(comando string, parametro string) {
	aux := comando + parametro
	s.serverlog = append(s.serverlog, aux)
	addtime, _ := strconv.Atoi(s.reloj[s.serverID])
	addtime++
	s.reloj[s.serverID] = strconv.Itoa(addtime)
}

func (s *server) propagarCambios(client pb.FulcrumClient) error {
	Add := []string{}
	var flag int
	for _, lines := range s.serverlog {
		parts := strings.Fields(lines)
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
	_, err := client.Propagar(context.Background(), &emptypb.Empty{})
	if err != nil {
		return err
	}

	return nil
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

func (s *server) sincronizarConDominante(client pb.FulcrumClient, reloj [3]string, log []string) error {
	res, err := client.Sincronizar(context.Background(), &pb.SyncRequest{
		RelojVector: reloj[:],
		Logs:        log,
	})
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}
	copy(reloj[:], res.RelojVector[:3])
	s.reloj = reloj
	s.serverlog = res.Logs
	return nil
}

func main() {
	serverID := 2
	reloj := [3]string{"0", "0", "0"}
	serverlog := []string{}

	s := &server{
		serverID:  serverID,
		reloj:     reloj,
		serverlog: serverlog,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051+serverID))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFulcrumServer(grpcServer, s)
	log.Printf("Server %d listening at %v", serverID, lis.Addr())

	conn, err := grpc.Dial("localhost:50061", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar con el server dominante: %v", err)
	}

	client := pb.NewFulcrumClient(conn)

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			err = s.sincronizarConDominante(client, reloj, serverlog)
			if err != nil {
				log.Printf("Error syncing with dominant server: %v", err)
			} else {
				err := s.propagarCambios(client)
				if err != nil {
					log.Printf("Error propagating changes: %v", err)
				} else {
					fmt.Println("Changes propagated successfully")
				}
			}
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	defer conn.Close()
	select {}

}
