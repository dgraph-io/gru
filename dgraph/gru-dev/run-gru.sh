#!/bin/sh

sleep 5
echo pwd
echo ~/
ls ~/.ssh
apt-get install openssh
sudo ps ax
echo GO GO GO SSH
ssh -L 11001:localhost:8080 -L 11002:localhost:9080 Gru &
sleep 10
echo GO GO GO GRU
/gru --dgraph=localhost:11002 --httpdgraph=http://localhost:11001 --ip="http://localhost" --user=gru-uz --pass=gru-pwd --secret=long-skt --sendgrid="$SENDGRID_API_KEY" --debug=true
