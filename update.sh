#!/bin/sh

docker-compose build --parallel
docker-compose push
kubectl apply -f k8s.yaml