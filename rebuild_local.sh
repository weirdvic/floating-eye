#!/bin/env bash

CONTAINER_RUNNER="docker"
IMAGE_NAME="floating-eye"

${CONTAINER_RUNNER} stop ${IMAGE_NAME}
${CONTAINER_RUNNER} rm ${IMAGE_NAME}
${CONTAINER_RUNNER} build -t ${IMAGE_NAME} .
${CONTAINER_RUNNER} run --name=${IMAGE_NAME} \
        --detach \
        --restart=always \
        ${IMAGE_NAME}:latest
