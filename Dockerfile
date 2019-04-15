FROM golang:1.12 as builder

ARG GOBIN=/go/bin/
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0

ENV PKG_NAME=github.com/DeFacto-Team/Factom-Open-API
ENV PKG_PATH=$GOPATH/src/$PKG_NAME

WORKDIR $PKG_PATH
COPY glide.yaml glide.lock $PKG_PATH/

RUN curl https://glide.sh/get | sh && \
    glide install -v

COPY . $PKG_PATH/

RUN go build -o /go/bin/factom-open-api main.go
RUN go build -o /go/bin/user admin/user.go

FROM alpine:3.7

RUN apk add --no-cache ca-certificates curl git

COPY --from=builder /go/bin/factom-open-api /factom-open-api
COPY --from=builder /go/bin/user /user
COPY --from=builder /go/src/github.com/DeFacto-Team/Factom-Open-API/migrations /migrations

CMD ["/factom-open-api"]