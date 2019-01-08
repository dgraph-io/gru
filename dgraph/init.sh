# docker run -v vol-gru-serverdata:/data --name helper busybox true
# docker cp . helper:/data
# docker rm helper

sudo mkdir -p /var/log/nginx
sudo mkdir -p /home/ubuntu/gru/server-data/

go build ..
