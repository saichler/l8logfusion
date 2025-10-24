#!/usr/bin/env bash
set -e
docker build --no-cache --platform=linux/amd64 -t saichler/probler-logagent:latest .
docker push saichler/probler-logagent:latest
