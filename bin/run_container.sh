#!/usr/bin/env bash

CONTAINER_NAME='foa'
PORT='8081'
POSTGRES_CONTAINER_NAME='foa-db'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

cd ${SCRIPT_DIR}/..

docker stop ${CONTAINER_NAME}

docker rm ${CONTAINER_NAME}

docker run \
  -d \
  -p ${PORT}:${PORT} \
  -v ${SCRIPT_DIR}/../testing:/home/app/values \
  --link ${POSTGRES_CONTAINER_NAME}:${POSTGRES_CONTAINER_NAME} \
  --name ${CONTAINER_NAME} \
  factom-open-api:latest
  