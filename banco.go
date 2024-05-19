package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	proto "github.com/FranciscoMunozP/Lab4Proto"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedDoshBankServer
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func (s *server) GetCurrentBalance(ctx context.Context, req *proto.PrepareRequest) (*proto.MoneyBalance, error) {
	// Abrir el archivo
	file, err := os.Open("datos.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Crear un lector de bufio
	scanner := bufio.NewScanner(file)

	// La cadena que estamos buscando
	searchString := fmt.Sprintf("Mercemario %s", req.MercenaryId)
	var line string
	// Leer el archivo línea por línea
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, searchString) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Retornar el contenido leído como parte de la respuesta
	return &proto.MoneyBalance{Balance: line}, nil
}

func main() {
	//Crear archivo
	file, err := os.OpenFile("datos.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	failOnError(err, "Error al abrir el archivo")
	defer file.Close()

	writer := bufio.NewWriter(file)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/") // Conexión a RabbitMQ
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel() // Crear un canal
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"eliminations", // Nombre de la cola
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}
	// Registro del servidor DoshBank
	lis, err := net.Listen("tcp", ":50054")
	failOnError(err, "Failed to listen")
	s := grpc.NewServer()
	proto.RegisterDoshBankServer(s, &server{})

	// Iniciar el servidor DoshBank
	log.Println("DoshBank server started on port 50054")
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	go func() {
		d, ok := <-msgs
		_ = ok
		parts := strings.Split(string(d.Body), ",") // Convert d.Body to string
		_, err = writer.WriteString(fmt.Sprintf("Mercenario %s piso %s acumulado actual %s;\n", parts[0], parts[1], parts[2]))
		failOnError(err, "Error al escribir en el archivo")

		// Flushear el escritor para asegurarse de que todos los datos se escriban en el archivo
		err = writer.Flush()
		failOnError(err, "Error al flushear el escritor")

		// Si no hubo errores, imprimir mensaje de éxito
		log.Printf("Dato \" %s \" recibido y escrito en el archivo correctamente.", d.Body)
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
