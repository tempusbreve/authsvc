#!/usr/bin/env bash

if [ -f ".local.config" ]; then
  . ".local.config"
fi


IMAGE=tempusbreve/authsvc
NAME="${NAME:-authsvc}"
PORT=${PORT:-4884}
VERBOSE=${VERBOSE:-}

SCRIPTDIR="$(cd "$(dirname "$0")"; pwd -P)"

function pull-image {
  docker pull ${IMAGE}
}

function start-container {
  DATA="${SCRIPTDIR}/data/"
  [ -d "${DATA}" ] || mkdir -p "${DATA}"

  docker run -d \
      --name ${NAME} \
      --restart=always \
      -e VERBOSE="${VERBOSE}" \
      -e PORT="${PORT}" \
      -p ${PORT}:${PORT} \
      -v "${DATA}":/data \
    ${IMAGE}
}

function kill-container {
  LOGS="${SCRIPTDIR}/logs/"
  [ -d "${LOGS}" ] || mkdir -p "${LOGS}"

  docker stop ${NAME} 2> /dev/null && \
    docker logs ${NAME} &> ${LOGS}$(TZ=UTC date +%Y-%m-%d-%H%M-${NAME}.log) && \
    docker rm -v -f ${NAME}
}

function log-container {
  docker logs ${NAME}
}

function log-follow-container {
  docker logs -f ${NAME}
}

function log-tail-container {
  docker logs --tail ${1:-30} ${NAME}
}

