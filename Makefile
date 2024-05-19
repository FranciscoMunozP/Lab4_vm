RabittMQ:
	sudo docker pull rabbitmq
	sudo docker run -d --name rabbitmq --network mynetwork -p 15672:15672 -p 5672:5672 rabbitmq:latest

network:
	sudo docker network create mynetwork

docker-Director:
	sudo docker build -f Dockerfile.Director . -t director:latest
	sudo docker run -d --name director --network mynetwork -p 50051:50051 director:latest


clean:
	# sudo docker network rm mynetwork
	# sudo docker stop rabbitmq
	# sudo docker rm rabbitmq


start-all: clean network RabittMQ docker-Director
	
