FROM alpine:latest
#FROM hypriot/rpi-alpine-scratch

COPY core-networking /go/bin/
WORKDIR /go/bin
CMD ["./core-networking"]
