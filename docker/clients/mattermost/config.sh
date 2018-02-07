#!/usr/bin/env bash

if [ -f ".local.config" ]; then
  . ".local.config"
fi


IMAGE=mattermost/mattermost-preview
NAME="${NAME:-mattermost}"
PORT=${PORT:-8065}

function pull-image {
  docker pull ${IMAGE}
}

function start-container {
  docker run -d \
      --name ${NAME} \
      --restart=always \
      -e PORT="${PORT}" \
      -p ${PORT}:${PORT} \
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
