package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	pb "github.com/FranciscoMunozP/Lab4Proto"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

var (
	isded bool = false
	mu1   sync.Mutex
	mu2   sync.Mutex
	mu3   sync.Mutex
	mu4   sync.Mutex
	wg    sync.WaitGroup
)

// const players int = 1
const ipNameNode string = "localhost:50053"
const ipDoshBank string = "localhost:50054"

// DirectorService implementa el servicio Director
type DirectorService struct {
	pb.UnimplementedDirectorServer
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (s *DirectorService) CheckMoneyBalance(ctx context.Context, req *pb.PrepareRequest) (*pb.MoneyBalance, error) {
	conn, err := grpc.Dial(ipDoshBank, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to DoshBank: %v", err)
	}
	mercenario := pb.NewDoshBankClient(conn)

	balance, err := mercenario.GetCurrentBalance(context.Background(), &pb.PrepareRequest{MercenaryId: req.MercenaryId})
	if err != nil {
		log.Printf("Failed to call service method: %v", err)
	}
	conn.Close()
	return &pb.MoneyBalance{Balance: balance.Balance}, nil //fix later balance
}

func generarNumeros() (int, int) {
	x := rand.Intn(101)
	y := rand.Intn(101)
	for x == y {
		y = rand.Intn(101)
	}
	if x > y {
		x, y = y, x
	}
	return x, y
}

func main() {
	wg.Add(10)
	rand.Seed(time.Now().UnixNano())

	// Iniciar el servidor Director
	lis, err := net.Listen("tcp", ":50051")
	failOnError(err, "failed to listen")
	s := grpc.NewServer()

	pb.RegisterDirectorServer(s, &DirectorService{})
	// pb.RegisterMercenarioServer(s, &MercenarioService{})

	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	//----Conecciones----
	//--Conectar con RabbitMQ--
	RabbitMQ, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer RabbitMQ.Close()

	ch, err := RabbitMQ.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	//Decleracion de la cola
	q, err := ch.QueueDeclare(
		"eliminations", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")
	defer RabbitMQ.Close()
	//--Conectar con el NameNode--
	conn2, err := grpc.Dial(ipNameNode, grpc.WithInsecure())
	failOnError(err, "Failed to connect to NameNode")
	NameNode := pb.NewNameNodeClient(conn2)

	time.Sleep(2 * time.Second)
	wg.Add(2)
	// Iniciar la corutina del juego

	for i := 0; i < 2; i++ {
		go func(id int) {
			defer wg.Done()
			// inicializar la semilla de los números aleatorios
			rand.Seed(time.Now().UnixNano())
			mu1.Lock()
			//--Conectar con el Mercenario--
			var conn1 *grpc.ClientConn
			var err error
			if id == 0 {
				conn1, err = grpc.Dial("localhost:50052", grpc.WithInsecure())

			} else {
				conn1, err = grpc.Dial("localhost:50056", grpc.WithInsecure())
			}
			failOnError(err, "Failed to connect to Mercenario")
			mercenario := pb.NewMercenarioClient(conn1)
			// Preparar al mercenario enviando su ID
			_, err = mercenario.ReportPrepareness(context.Background(), &pb.PrepareRequest{MercenaryId: strconv.Itoa(id)})
			failOnError(err, "Failed to call service method")
			defer conn1.Close()

			mu1.Unlock()
			mu2.Lock()
			//----Piso 1----
			choice, err := mercenario.ChooseProcess(context.Background(), &pb.ChoiceRequest{MercenaryId: strconv.Itoa(id), Floor: "1"})
			failOnError(err, "Failed to call service method")
			_, err = NameNode.StoreChoice(context.Background(), &pb.ChoiceSave{MercenaryId: choice.MercenaryId, Floor: "1", Choice: choice.Choice})
			failOnError(err, "Failed to call service method")
			x, y := generarNumeros()
			probabilidades := []int{x, y - x, 100 - y}
			choiceIndex, _ := strconv.Atoi(choice.Choice)
			probabilidad := probabilidades[choiceIndex]
			if !(rand.Intn(100) < probabilidad) {
				_, err := mercenario.NotifyElimination(context.Background(), &pb.EliminationNotification{MercenaryId: strconv.Itoa(id), Alert: "Dead"})
				failOnError(err, "Failed to call service method")
				// Enviar mensaje a la cola de eliminaciones
				body := fmt.Sprintf("%s,%s,%s", choice.MercenaryId, "1", choice.Choice)
				err = ch.Publish(
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(body),
					})
				failOnError(err, "Failed to publish a message")
				log.Printf(" [x] Sent %s", body) //NotyfyElimination
				isded = true
				return
			} else {
				fmt.Printf("Mercenario %s ha sobrevivido al Piso 1.\n", strconv.Itoa(id))
			}
			mu2.Unlock()
			mu3.Lock()
			//Piso 2
			if !isded {
				choice, err = mercenario.ChooseProcess(context.Background(), &pb.ChoiceRequest{MercenaryId: "1", Floor: "2"})
				failOnError(err, "Failed to call service method")
				_, err = NameNode.StoreChoice(context.Background(), &pb.ChoiceSave{MercenaryId: choice.MercenaryId, Floor: "2", Choice: choice.Choice})
				failOnError(err, "Failed to call service method")
				opcionCorrecta := strconv.Itoa(rand.Intn(2))
				if !(choice.Choice == opcionCorrecta) {
					_, err := mercenario.NotifyElimination(context.Background(), &pb.EliminationNotification{MercenaryId: "1", Alert: "Dead"})
					failOnError(err, "Failed to call service method")
					// Enviar mensaje a la cola de eliminaciones
					body := fmt.Sprintf("%s,%s,%s", choice.MercenaryId, "2", choice.Choice)
					err = ch.Publish(
						"",     // exchange
						q.Name, // routing key
						false,  // mandatory
						false,  // immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        []byte(body),
						})
					failOnError(err, "Failed to publish a message")
					log.Printf(" [x] Sent %s", body) //NotyfyElimination
					isded = true
					return
				} else {
					fmt.Printf("Mercenario %s ha sobrevivido al Piso 2.\n", strconv.Itoa(id))
				}
				mu3.Unlock()
				mu4.Lock()
				//Piso 3
				if !isded {
					aciertos := 0
					for i := 0; i < 5; i++ {
						numeroPatriarca := strconv.Itoa(rand.Intn(15) + 1)
						choice, err = mercenario.ChooseProcess(context.Background(), &pb.ChoiceRequest{MercenaryId: "1", Floor: fmt.Sprintf("3,%d", i)})
						failOnError(err, "Failed to call service method")
						_, err = NameNode.StoreChoice(context.Background(), &pb.ChoiceSave{MercenaryId: choice.MercenaryId, Floor: "3", Choice: choice.Choice})
						failOnError(err, "Failed to call service method")
						if numeroPatriarca == choice.Choice {
							aciertos++
						}
						//fmt.Printf("Número del Patriarca: %s\n", numeroPatriarca)
					}

					if aciertos >= 2 {
						fmt.Printf("Mercenario %s ha vencido al Patriarca y completado la misión.\n", strconv.Itoa(id))
					} else {
						_, err := mercenario.NotifyElimination(context.Background(), &pb.EliminationNotification{MercenaryId: "1", Alert: "Dead"})
						failOnError(err, "Failed to call service method")
						// Enviar mensaje a la cola de eliminaciones
						body := fmt.Sprintf("%s,%s,%s", choice.MercenaryId, "3", choice.Choice)
						err = ch.Publish(
							"",     // exchange
							q.Name, // routing key
							false,  // mandatory
							false,  // immediate
							amqp.Publishing{
								ContentType: "text/plain",
								Body:        []byte(body),
							})
						failOnError(err, "Failed to publish a message")
						log.Printf(" [x] Sent %s", body) //NotyfyElimination
						isded = true
						return
					}
				}
				mu4.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if isded {
		log.Printf("gg no re")
	} else {
		log.Printf("gg ez")
	}
}
