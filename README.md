# Gru

Gru is an open source adaptive test system for screening candidates for software engineering roles. It helps us identify and recruit the right minions.

You can read more about why we built Gru on our [blog](https://open.dgraph.io/post/gru/). Gru uses Dgraph as the database.

## Instructions

1. Get the Gru code and checkout the develop branch.
  - go get github.com/dgraph-io/gru
  - cd $GOPATH/src/github.com/dgraph-io/gru
  - git checkout develop
2. Start Dgraph on port 8080. Note we are currently building and using Dgraph from the master right now.
So you can build Dgraph from source or ask us for the Dgraph binary for Linux/Mac.
3. Start the admin backend from gru directory like `go build . && ./gru -user=admin -pass=pass -secret=0a45e5eGseF41o0719PJ39KljMK4F4v2`.
Note we use sendgrid for sending invite mails to the candidates. That won't work without the sendgrid
key. For development purposes when the `--sendgrid` flag is empty, we just print out the invite link for taking
the quiz to the console.
4. We use [Caddy](https://caddyserver.com/) as the Web server and run it on port 2020 locally. You can install
caddy and then run it from inside admin/webUI using `caddy` command.
You can checkout the Caddyfile in the folder for config for Caddy. The easiest method to install caddy is using
https://getcaddy.com/.

After this Gru should be up and running for you. You can visit http://localhost:2020 and use
`admin` as username and `pass` as password for logging in. Go ahead and add some questions, create some quizzes and invite some candidates.

# [![Coverage Status](https://coveralls.io/repos/github/dgraph-io/gru/badge.svg?branch=develop)](https://coveralls.io/github/dgraph-io/gru?branch=develop)
