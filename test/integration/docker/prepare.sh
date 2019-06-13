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
