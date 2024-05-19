network:
	sudo docker network create mynetwork


docker-NameNode:
	sudo docker build -f Dockerfile.NameNode . -t namenode:latest
	sudo docker run --rm --name NameNode --network="host" namenode:latest

clean:
	sudo docker network rm mynetwork

start-all: clean network 
	
