## Build binary
bin:
	go build  --ldflags '-linkmode "external" -extldflags "-static"' -x -o core-networking .
## Build with version number for test purposes
build:
	sudo docker build -t iofog/core-networking:$(TAG) .
## Push with version number for test purposes
push:
	sudo docker push iofog/core-networking:$(TAG)
## Tag latest to verified version number
latest:
	sudo docker tag iofog/core-networking:$(TAG) iofog/core-networking
## Push latest
push-latest:
	sudo docker push iofog/core-networking

## Same cmds for arm
build-arm:
	sudo docker build -t iofog/core-networking-arm:$(TAG) -f Dockerfile-arm .
push-arm:
	sudo docker push iofog/core-networking-arm:$(TAG)
latest-arm:
	sudo docker tag iofog/core-networking-arm:$(TAG) iofog/core-networking-arm
push-latest-arm:
	sudo docker push iofog/core-networking-arm
