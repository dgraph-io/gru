# Gru

Gru is an open source adaptive test system for screening candidates for software engineering roles. It helps us identify and recruit the right minions.

To understand how Gru works, head over to our [Wiki](https://wiki.dgraph.io/Gru).

You can read more about why we built Gru on our [blog](https://open.dgraph.io/post/gru/). Gru uses Dgraph as the database.

## Instructions

1. Start Dgraph on port 8080.
2. Start the admin backend from inside `gruadmin` directory like `go build . && ./gruadmin`.
3. The frontend is built using AngularJS. To run the admin panel run `python -m SimpleHTTPServer` from `gruadmin/webUI` directory. Now you can go to `127.0.0.1:8000` and access the admin panel.

# [![Coverage Status](https://coveralls.io/repos/github/dgraph-io/gru/badge.svg?branch=develop)](https://coveralls.io/github/dgraph-io/gru?branch=develop)
