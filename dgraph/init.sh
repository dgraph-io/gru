
docker volume create vol-gru-dgraph
docker volume create vol-gru-serverdata

# docker run -v vol-gru-serverdata:/data --name helper busybox true
# docker cp . helper:/data
# docker rm helper
