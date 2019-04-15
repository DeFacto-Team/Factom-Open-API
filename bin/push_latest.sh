#!/usr/bin/env bash

NAMESPACE='defactoteam'
IMAGE_NAME='factom-open-api'
TAG='latest'

set -xe

docker push ${NAMESPACE}/${IMAGE_NAME}:${TAG}
