#!/bin/sh

sleep 5
/gru --dgraph=gru-alpha:9080 --httpdgraph=http://gru-alpha:8080 --ip="http://localhost" --user=gru-uz --pass=gru-pwd --secret=long-skt --sendgrid="$SENDGRID_API_KEY" --debug=true
