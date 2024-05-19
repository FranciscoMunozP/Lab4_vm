package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"

	proto "github.com/FranciscoMunozP/Lab4Proto"
)

const ip_dt1 string = "localhost:50051"
const ip_dt2 string = "localhost:50052"
const ip_dt3 string = "localhost:50053"

var Conexion = make(map[string]string)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type server struct {
	proto.UnimplementedNameNodeServer
}

func RandDistribution() {
	rand.Seed(time.Now().UnixNano()) // Inicializar la semilla del generador de números aleatorios

	numeros := make([]int, 0, 8)      // Crear un slice para almacenar los números
	yaGenerados := make(map[int]bool) // Crear un mapa para llevar un registro de los números ya generados

	for len(numeros) < 8 {
		num := rand.Intn(8) + 1 // Generar un número aleatorio entre 1 y 8
		if !yaGenerados[num] {
			numeros = append(numeros, num) // Añadir el número al slice
			yaGenerados[num] = true        // Marcar el número como ya generado
		}
	}
	Conexion[strconv.Itoa(numeros[0])] = ip_dt1
	Conexion[strconv.Itoa(numeros[1])] = ip_dt1
	Conexion[strconv.Itoa(numeros[2])] = ip_dt2
	Conexion[strconv.Itoa(numeros[3])] = ip_dt2
	Conexion[strconv.Itoa(numeros[4])] = ip_dt2
	Conexion[strconv.Itoa(numeros[5])] = ip_dt3
	Conexion[strconv.Itoa(numeros[6])] = ip_dt3
	Conexion[strconv.Itoa(numeros[7])] = ip_dt3

}

func (s *server) StoreChoice(ctx context.Context, req *proto.ChoiceSave) (*proto.Confirmation, error) {
	conn1, err := grpc.Dial(Conexion[req.MercenaryId], grpc.WithInsecure())
	failOnError(err, "Failed to connect to DataNode in STORE CHOICE")
	DataNode := proto.NewDataNodeClient(conn1)
	_, err = DataNode.StoreToFile(context.Background(), &proto.ChoiceSave{MercenaryId: req.MercenaryId, Floor: req.Floor, Choice: req.Choice})
	if err != nil {
		log.Fatalf("Failed to call service method: %s", err)
		return &proto.Confirmation{Responde: "Error"}, nil
	}
	return &proto.Confirmation{Responde: "OK"}, nil
}

func (s *server) RecoverChoice(ctx context.Context, req *proto.ChoiceRequest) (*proto.ChoiceMercenary, error) {
	conn1, err := grpc.Dial(Conexion[req.MercenaryId], grpc.WithInsecure())
	failOnError(err, "Failed to connect to DataNode in RECOVER CHOICE")
	DataNode := proto.NewDataNodeClient(conn1)
	choice, err := DataNode.RequestChoice(context.Background(), &proto.ChoiceRequest{MercenaryId: req.MercenaryId, Floor: req.Floor})
	if err != nil {
		log.Fatalf("Failed to call service method: %s", err)
	}
	return &proto.ChoiceMercenary{MercenaryId: choice.MercenaryId, Choice: choice.Choice}, nil
}
func main() {
	RandDistribution()
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterNameNodeServer(s, &server{})
	fmt.Println("NameNode server started...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	select {}
}
