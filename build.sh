#!/bin/bash
set -ev
cd backend
go test
go install
cd ../mongofs
go test
go install
cd ..
go test
GO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out/main .
docker build -t openservice/media-service:latest .
if [ "${TRAVIS_PULL_REQUEST}" = "false" ] && [ "${TRAVIS_REPO_SLUG}" = "InteractiveLecture/media-service" ] ; then
  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD" -e="$DOCKER_EMAIL"
  docker push openservice/media-service:latest
fi
