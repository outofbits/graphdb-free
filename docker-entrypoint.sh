#!/bin/sh
set -e

USER="${GDB_USER:=root}"

if [ "$USER" != "root" -a "$USER" != "0" ]; then
  if [ $(expr "$USER" : "^[0-9]*$") -eq 0 ]; then
    USER_PWD=$(cat /etc/passwd | grep -e "^$USER:" || true)
    if [ -z "$USER_PWD" ]; then
      adduser --no-create-home --disabled-password --gecos '' $USER
    fi
  fi
  set-ownership $USER
fi

mkdir -p /repository.init
mkdir -p /tmp/graphdb/toLoad

exec tini -g gosu -- $USER run-graphdb "$@"