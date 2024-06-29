package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/FranciscoMunozP/Lab5_proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ingeniero struct {
	cambios    []string
	ClientName string
}

func main() {
	cambios := []string{}

	i := &ingeniero{
		cambios:    cambios,
		ClientName: "Malkor",
	}

	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to Broker: %v", err)
	}
	defer conn.Close()

	brokerClient := pb.NewBrokerClient(conn)
	reader := bufio.NewReader(os.Stdin)
	var fulcrumClient pb.FulcrumClient
	var serverConn *grpc.ClientConn

	for {
		fmt.Print("\nC:\\Users\\" + i.ClientName + "> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Cerrando conexión y terminando sessión...")
			break
		}

		parts := strings.Fields(input)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if fulcrumClient == nil || serverConn == nil {
			serverResp, err := brokerClient.GetServer(ctx, &emptypb.Empty{})
			if err != nil {
				log.Printf("Failed to get server from Broker: %v", err)
				continue
			}

			serverConn, err = grpc.Dial(serverResp.ServerAddress, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Printf("Failed to connect to server: %v", err)
				continue
			}
			fulcrumClient = pb.NewFulcrumClient(serverConn)
		}

		switch parts[0] {
		case "AgregarBase":
			if len(parts) < 3 {
				fmt.Println("Error en sintaxis, trata AgregarBase <sector> <base> [cantidad]")
				continue
			}
			sector := parts[1]
			base := parts[2]
			cantidad := "0"
			if len(parts) == 4 {
				cantidad = parts[3]
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := fulcrumClient.AgregarBase(ctx, &pb.DatosSector{
				NombreSector:     sector,
				NombreBase:       base,
				CantidadEnemigos: cantidad,
			})
			cancel()
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Base creada con éxito")
				log.Printf("Reloj: [" + resp.RelojVector[0] + "," + resp.RelojVector[1] + "," + resp.RelojVector[2] + "]")
				i.cambios = append(i.cambios, parts[0]+" "+sector+" "+base+" "+cantidad)
			}

		case "RenombrarBase":
			if len(parts) < 4 {
				fmt.Println("Error en sintaxis, trata: RenombrarBase <sector> <base> <nueva_base>")
				continue
			}
			sector := parts[1]
			base := parts[2]
			nuevaBase := parts[3]

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := fulcrumClient.RenombrarBase(ctx, &pb.DatosSectorRenombrar{
				NombreSector:      sector,
				NombreBaseAntigua: base,
				NombreBaseNueva:   nuevaBase,
			})
			cancel()
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Base renombrada con éxito")
				log.Printf("Reloj: [" + resp.RelojVector[0] + "," + resp.RelojVector[1] + "," + resp.RelojVector[2] + "]")
				i.cambios = append(i.cambios, parts[0]+" "+sector+" "+base+" "+nuevaBase)
			}

		case "ActualizarValor":
			if len(parts) < 4 {
				fmt.Println("Error en sintaxis, trata: ActualizarValor <sector> <base> <cantidad>")
				continue
			}
			sector := parts[1]
			base := parts[2]
			cantidad := parts[3]

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := fulcrumClient.ActualizarValor(ctx, &pb.DatosSectorActualizar{
				NombreSector:     sector,
				NombreBase:       base,
				CantidadEnemigos: cantidad,
			})
			cancel()
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Cantidad de enemigos actualizada con éxito")
				log.Printf("Reloj: [" + resp.RelojVector[0] + "," + resp.RelojVector[1] + "," + resp.RelojVector[2] + "]")
				i.cambios = append(i.cambios, parts[0]+" "+sector+" "+base+" "+cantidad)
			}

		case "BorrarBase":
			if len(parts) < 3 {
				fmt.Println("Error en sintaxis, trata: BorrarBase <sector> <base>")
				continue
			}
			sector := parts[1]
			base := parts[2]

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := fulcrumClient.BorrarBase(ctx, &pb.DatosSectorConsulta{
				NombreSector: sector,
				NombreBase:   base,
			})
			cancel()
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Base borrada con éxito")
				log.Printf("Reloj: [" + resp.RelojVector[0] + "," + resp.RelojVector[1] + "," + resp.RelojVector[2] + "]")
				i.cambios = append(i.cambios, parts[0]+" "+sector+" "+base)
			}
		case "Log":
			for _, line := range i.cambios {
				log.Println(line)
			}
		default:
			log.Printf("Error: Comando Desconocido")
		}
	}

	if serverConn != nil {
		serverConn.Close()
	}
}
