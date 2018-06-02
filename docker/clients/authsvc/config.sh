#!/usr/bin/env bash

if [ -f ".local.config" ]; then
  . ".local.config"
fi


IMAGE=tempusbreve/authsvc
NAME="${NAME:-authsvc}"
PORT=${PORT:-4884}
BIND_IP=${BIND_IP:-0.0.0.0}

DEBUG=${DEBUG:-}
CRYPT_HASH=${CRYPT_HASH:-}
CRYPT_BLOCK=${CRYPT_BLOCK:-}
CACHE_DIR=${CACHE_DIR:-/data}
PUBLIC_HOME=${PUBLIC_HOME:-/public}
STORAGE=${STORAGE:-boltdb}

LDAP_HOST=${LDAP_HOST:-localhost}
LDAP_PORT=${LDAP_PORT:-389}
LDAP_TLS=${LDAP_TLS:-false}
LDAP_ADMIN_USER=${LDAP_ADMIN_USER:-cn=admin,dc=example,dc.com}
LDAP_ADMIN_PASS=${LDAP_ADMIN_PASS:-password}
LDAP_BASE_DN=${LDAP_BASE_DN:-dc=example,dc=com}



SCRIPTDIR="$(cd "$(dirname "$0")"; pwd -P)"

function pull-image {
  docker pull ${IMAGE}
}

function start-container {
  DATA_DIR="${SCRIPTDIR}/data/"
  [ -d "${DATA_DIR}" ] || mkdir -p "${DATA_DIR}"

  docker run -d \
      --name ${NAME} \
      --restart=always \
      -e DEBUG="${DEBUG}" \
      -e PORT="${PORT}" \
      -e BIND_IP="${BIND_IP}" \
      -e CRYPT_HASH="${CRYPT_HASH}" \
      -e CRYPT_BLOCK="${CRYPT_BLOCK}" \
      -e CACHE_DIR="${CACHE_DIR}" \
      -e PUBLIC_HOME="${PUBLIC_HOME}" \
      -e STORAGE="${STORAGE}" \
      -e LDAP_HOST="${LDAP_HOST}" \
      -e LDAP_PORT="${LDAP_PORT}" \
      -e LDAP_TLS="${LDAP_TLS}" \
      -e LDAP_ADMIN_USER="${LDAP_ADMIN_USER}" \
      -e LDAP_ADMIN_PASS="${LDAP_ADMIN_PASS}" \
      -e LDAP_BASE_DN="${LDAP_BASE_DN}" \
      -p ${PORT}:${PORT} \
      -v "${DATA_DIR}":/data \
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

