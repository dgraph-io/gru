cd .. && GOOS=linux GOARCH=386 time go build . && cd dgraph && docker-compose -f compose-local.yml down && docker-compose -f compose-local.yml up -d && docker ps -a | grep run-gru.sh
