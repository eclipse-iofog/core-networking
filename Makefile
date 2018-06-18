bin:
	go build  --ldflags '-linkmode "external" -extldflags "-static"' -x -o core-networking .
build:
	sudo docker build -t iofog/core-networking$(TAG) .
push:build
	sudo docker push iofog/core-networking-noack-go$(TAG)
