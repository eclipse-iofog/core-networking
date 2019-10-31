FROM golang:1.12-alpine as builder

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

COPY . /go/src/github.com/eclipse-iofog/core-networking
WORKDIR /go/src/github.com/eclipse-iofog/core-networking

RUN apk add --update --no-cache bash curl git make
RUN make vendor
RUN . version && export MAJOR && export MINOR && export PATCH && export SUFFIX && make build
RUN cp bin/cn /usr/bin/cn

FROM scratch
COPY --from=builder /usr/bin/cn /usr/bin/cn

CMD [ "/usr/bin/cn" ]