#!/bin/bash

apt-get update
apt-get upgrade -y

apt-get install -y docker.io docker-compose-v2 git

usermod -aG docker ubuntu

iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X
iptables -t mangle -F
iptables -t mangle -X
iptables -P INPUT ACCEPT
iptables -P FORWARD ACCEPT
iptables -P OUTPUT ACCEPT

apt-get install -y iptables-persistent
netfilter-persistent save

systemctl enable docker
systemctl start docker

echo "Cloud-init успішно завершено! Docker готовий до роботи."