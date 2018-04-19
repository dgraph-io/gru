# Gru

Gru is an open source adaptive test system for screening candidates for software engineering roles. It helps us identify and recruit the right minions.

You can read more about why we built Gru on our [blog](https://open.dgraph.io/post/gru/). Gru uses Dgraph as the database.

## Running

Gru has three components, Gru server, Dgraph(v0.7.5) as the database and Caddy as a web server.

### Gru server

```
  # Make sure you have Go installed on the server (https://golang.org/doc/install).
  go get github.com/dgraph-io/gru
  cd $GOPATH/src/github.com/dgraph-io/gru
  go build .
  ./gru --user=<gru_ui_username> --pass=<gru_ui_password> --secret="<long_secret_to_sign_jwt_token>" --sendgrid="<sendgrid_api_key>" --gh '<greenhouse_api_key>' -ip "https://gru.dgraph.io" -debug=true 2>&1 | tee -a gru.log
```

* Note we use sendgrid for sending invite mails to the candidates. That won't work without the sendgrid
key. For development purposes when the `--sendgrid` flag is empty, we just print out the invite link for taking
the quiz to the console.
* Gru also has an integration with Greenhouse and the api key can be supplied using the `gh` flag.
* The `-ip` flag is used to specify the ip address of the Gru server.


### Dgraph

```
  wget https://github.com/dgraph-io/dgraph/releases/download/v0.7.5/dgraph-linux-amd64-v0.7.5.tar.gz
  tar -xzvf dgraph-linux-amd64-v0.7.5.tar.gz
  ./dgraph/dgraph
```
In case you are reloading data into Dgraph from an export, you can use dgraphloader to load the `rdf.gz` exported file.

Dgraph runs on port 8080 by default.

### Caddy

Note, you should modify the the address on the first line of the `admin/webUI/Caddyfile` and also the
value of `hostname` in `admin/webUI/app/app.module.js` to either `http://localhost:2020` for the purposes
of local development or to the address of your production server before running Caddy web server.

```
  mkdir caddy
  wget https://github.com/mholt/caddy/releases/download/v0.10.8/caddy_v0.10.8_linux_amd64.tar.gz
  tar -xzvf caddy_v0.10.8_linux_amd64.tar.gz
  ./caddy --conf ../admin/webUI/Caddyfile
```

After this Gru should be up and running for you. You can visit http://localhost:2020 (if running
locally) and login. Go ahead and add some questions, create some quizzes and invite some candidates.

# [![Coverage Status](https://coveralls.io/repos/github/dgraph-io/gru/badge.svg?branch=develop)](https://coveralls.io/github/dgraph-io/gru?branch=develop)
