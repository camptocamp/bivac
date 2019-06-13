#!/bin/bash

echo "[*] Retrieving internal IP..."
export INTERNAL_IP=$(ip -o -4 addr show dev `ls /sys/class/net | grep -E "^eth|^en" | head -n 1` | cut -d' ' -f7 | cut -d'/' -f1)

echo "[*] Starting local registry..."
docker run --restart=always -d -p 5000:5000 --name registry registry:2
sleep 10

echo "[*] Starting Minio..."
docker run --restart=always -d -p 9000:9000 -e MINIO_ACCESS_KEY=OBQZY3DV6VOEZ9PG6NIM -e MINIO_SECRET_KEY=7e88XeX0j3YdB6b1o0zU2GhG0dX6tFMy3Haty --name minio -v /root/minio:/data minio/minio server /data
sleep 10
docker pull minio/mc
docker run --rm -e MC_HOST_minio=http://OBQZY3DV6VOEZ9PG6NIM:7e88XeX0j3YdB6b1o0zU2GhG0dX6tFMy3Haty@${INTERNAL_IP}:9000 minio/mc mb minio/bivac-testing

echo "[*] Starting Rancher..."
docker run -d --restart=unless-stopped -p 8080:8080 rancher/server:stable
sleep 60
curl 'http://localhost:8080/v2-beta/setting' -H 'Accept: application/json' -H 'content-type: application/json' --data '{"type":"setting","name":"telemetry.opt","value":"in"}'
sleep 1
curl 'http://localhost:8080/v2-beta/settings/api.host' -X PUT -H 'Accept: application/json' -H 'content-type: application/json' --data '{"id":"api.host","type":"activeSetting","baseType":"setting","name":"api.host","activeValue":null,"inDb":false,"source":null,"value":"http://'${INTERNAL_IP}':8080"}'
sleep 1
curl 'http://localhost:8080/v2-beta/projects/1a5/registrationtoken' --data '{"type":"registrationToken"}'
sleep 1
command=$(curl -s 'http://localhost:8080/v2-beta/projects/1a5/registrationtokens?state=active&limit=-1&sort=name'  -H 'Accept: application/json'  -H 'content-type: application/json' | jq -r ".data[0].command")
echo $command
$command

echo "[*] Installing Rancher CLI..."
wget https://releases.rancher.com/cli/v0.6.12/rancher-linux-amd64-v0.6.12.tar.gz
tar zxvf rancher-linux-amd64-v0.6.12.tar.gz
sudo cp ./rancher-v0.6.12/rancher /bin/rancher
sudo chmod +x /bin/rancher
rm -rf ./rancher-v0.6.12
rm rancher-linux-amd64-v0.6.12.tar.gz
rancher
sync
