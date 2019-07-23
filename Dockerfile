FROM node:12.4-alpine as uibuilder

WORKDIR /ui
COPY ./ui .

RUN npm install -g yarn
RUN yarn install && yarn build

FROM golang:1.12 AS builder

ARG GOBIN=/go/bin/
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0
ARG GO111MODULE=on
ARG PKG_NAME=github.com/DeFacto-Team/Factom-Open-API
ARG PKG_PATH=${GOPATH}/src/${PKG_NAME}

WORKDIR ${PKG_PATH}
COPY . ${PKG_PATH}/

RUN go mod download && \
  go build -o /go/bin/factom-open-api main.go

FROM alpine:3.7

RUN set -xe && \
  apk --no-cache add bash ca-certificates inotify-tools && \
  addgroup -g 1000 app && \
  adduser -D -G app -u 1000 app

WORKDIR /home/app

COPY --from=builder /go/bin/factom-open-api ./
COPY --from=uibuilder /ui/build ./ui/build
COPY ./entrypoint.sh ./entrypoint.sh
COPY ./migrations ./migrations
COPY ./docs/swagger.json ./docs/swagger.json

RUN \
  mkdir ./values && \
  chown -R app:app /home/app

USER app

EXPOSE 8081

ENTRYPOINT [ "./entrypoint.sh" ]

CMD [ "./factom-open-api", "-c", "/home/app/values/config.yaml" ]
