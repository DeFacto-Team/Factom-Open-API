FROM golang:1.12-alpine as builder

ENV SRC_DIR=/go/src/github.com/DeFacto-Team/Factom-Open-API/

RUN mkdir -p /root/.config
COPY .config/config.yaml /root/.config

ADD . $SRC_DIR
RUN cd $SRC_DIR; go build -o openapi; cp openapi /go/bin/

ADD openapi /go/bin

FROM alpine:3.7

RUN mkdir -p /root/.config /go/bin
COPY --from=builder /root/.config/config.yaml /root/.config/config.yaml
COPY --from=builder /go/bin/openapi /go/bin/openapi

ENTRYPOINT ["/go/bin/openapi"]

EXPOSE 8081