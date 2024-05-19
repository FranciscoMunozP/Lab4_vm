network:
	sudo docker network create mynetwork

docker-DataNode1:
	sudo docker build -f Dockerfile.Datanode-1 . -t datanode-1:latest
	sudo docker run --rm --name Datanode-1 -p 50052:50052 datanode-1:latest

docker-DataNode2:
	sudo docker build -f Dockerfile.Datanode-2 . -t datanode2:latest
	sudo docker run --rm --name Datanode-2 -p 50053:50053 datanode2:latest

docker-DataNode3:
	sudo docker build -f Dockerfile.Datanode-2 . -t datanode2:latest
	sudo docker run --rm --name Datanode-2 -p 50053:50053 datanode2:latest

clean:
	sudo docker network rm mynetwork
	rm -f ./datanode1/*.txt
	rm -f ./datanode2/*.txt
	rm -f ./datanode3/*.txt


start-all: clean network docker-DataNode1 docker-DataNode3 docker-DataNode3
	