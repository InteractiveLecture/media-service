#!/bin/bash
GO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out/main . && docker-compose rm -f && docker-compose build && docker-compose up
