package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/FranciscoMunozP/Lab4Proto" // Asegúrate de que la ruta sea correcta

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedDataNodeServer
}

func (s *server) RequestChoice(ctx context.Context, req *pb.ChoiceRequest) (*pb.ChoiceMercenary, error) {
	// Abrir el archivo
	file, err := os.Open(fmt.Sprintf("Mercenario%s_%s.txt", req.MercenaryId, req.Floor)) // Asegúrate de que la extensión del archivo sea correcta
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	// Crear un lector de bufio
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	file.Close()
	return &pb.ChoiceMercenary{MercenaryId: req.MercenaryId, Choice: line}, nil
}

func (s *server) StoreToFile(ctx context.Context, req *pb.ChoiceSave) (*pb.Confirmation, error) {
	log.Printf("Received decisions from Director for the mercenary %v", req.MercenaryId)

	// Abrir o crear el archivo para escritura
	file, err := os.OpenFile(fmt.Sprintf("Mercenario%s_%s.txt", req.MercenaryId, req.Floor), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	// Escribir los datos en el archivo
	if req.Floor == "1" {
		if req.Choice == "0" {
			_, err = file.WriteString("Mercenario " + req.MercenaryId + " escogió la Escopeta (" + req.Choice + ");\n")
		} else if req.Choice == "1" {
			_, err = file.WriteString("Mercenario " + req.MercenaryId + " escogió el Rifle automático (" + req.Choice + ");\n")
		} else if req.Choice == "2" {
			_, err = file.WriteString("Mercenario " + req.MercenaryId + " escogió los Puños eléctricos (" + req.Choice + ");\n")
		}
	} else if req.Floor == "2" {
		pasillo := ""
		if req.Choice == "0" {
			pasillo = "A"
		} else if req.Choice == "1" {
			pasillo = "B"
		}
		_, err = file.WriteString("Mercenario " + req.MercenaryId + " escogió el pasillo " + pasillo + " (" + req.Choice + ");\n")
	} else if req.Floor == "3" {
		_, err = file.WriteString("Mercenario " + req.MercenaryId + " escogió el número " + req.Choice + ";\n")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %v", err)
	}

	// Devolver una respuesta para indicar que se recibieron las decisiones
	file.Close()
	return &pb.Confirmation{Responde: "Decision saved successfully"}, nil
}

func main() {

	// Esperar un tiempo antes de la siguiente operación
	time.Sleep(time.Second * 5)

	// Iniciar el servidor DataNode
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDataNodeServer(s, &server{})

	log.Println("DataNode server started on port 50055")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	// Iniciar la rutina del cliente
	go func() {
		conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to NameNode: %v", err)
		}

		client := pb.NewNameNodeClient(conn)

		response, err := client.RecoverChoice(context.Background(), &pb.ChoiceRequest{MercenaryId: "some_id", Floor: "some_floor"})
		if err != nil {
			log.Printf("Failed to call service method: %v", err)
		} else {
			log.Printf("Received response from NameNode: %v", response)
		}
		select {}
	}()
	// Esperar indefinidamente

}
