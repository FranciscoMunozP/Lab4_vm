RabittMQ:
	sudo docker pull rabbitmq
	sudo docker run -d --name rabbitmq --network mynetwork -p 15672:15672 -p 5672:5672 rabbitmq:latest

network:
	sudo docker network create mynetwork

docker-DoshBank:
	sudo docker build -f Dockerfile.DoshBank  . -t doshbank:latest
	sudo docker run -d --name DoshBank --network mynetwork -p 50054:50054 doshbank:latest


clean:
	sudo docker network rm mynetwork
	# sudo docker stop rabbitmq
	# sudo docker rm rabbitmq


start-all: clean network RabittMQ docker-DoshBank