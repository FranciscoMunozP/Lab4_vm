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

func main() {
	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to Broker: %v", err)
	}
	defer conn.Close()

	brokerClient := pb.NewBrokerClient(conn)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nC:\\Users\\Kais> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Cerrando conexión y terminando programa...")
			break
		}

		parts := strings.Fields(input)
		if len(parts) != 3 || parts[0] != "GetEnemigos" {
			fmt.Println("Comando invalido")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		serverResp, err := brokerClient.GetServer(ctx, &emptypb.Empty{})
		if err != nil {
			log.Fatalf("Failed to get server from Broker: %v", err)
		}

		serverConn, err := grpc.Dial(serverResp.ServerAddress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Failed to connect to server: %v", err)
		}
		defer serverConn.Close()

		fulcrumClient := pb.NewFulcrumClient(serverConn)

		resp, err := fulcrumClient.GetEnemigos(ctx, &pb.DatosSectorConsulta{
			NombreSector: parts[1],
			NombreBase:   parts[2],
		})
		if err != nil {
			log.Fatalf("Failed to get number of enemies: %v", err)
		}
		fmt.Printf("La cantidad de enemigos en %s es %s\n", parts[2], resp.CantidadEnemigos)
	}
}
