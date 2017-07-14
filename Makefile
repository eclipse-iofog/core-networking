bin:
	go build  --ldflags '-linkmode "external" -extldflags "-static"' -x -o core-networking .
build:
	sudo docker build -t iotracks/catalog:core-networking-noack-go$(TAG) .
push:build
	sudo docker push iotracks/catalog:core-networking-noack-go$(TAG)
