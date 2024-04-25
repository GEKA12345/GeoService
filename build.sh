#!/bin/bash
go test -cover ./...
cd ./proxy
go generate
cd ../
docker-compose up --force-recreate --build
