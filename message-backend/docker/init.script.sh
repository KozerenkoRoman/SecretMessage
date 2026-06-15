#!/bin/bash

sudo apt-get update
sudo apt-get upgrade -y

sudo apt-get install -y docker.io docker-compose-v2 git

sudo usermod -aG docker ubuntu

sudo iptables -F
sudo iptables -X
sudo iptables -t nat -F
sudo iptables -t nat -X
sudo iptables -t mangle -F
sudo iptables -t mangle -X
sudo iptables -P INPUT ACCEPT
sudo iptables -P FORWARD ACCEPT
sudo iptables -P OUTPUT ACCEPT

sudo apt-get install -y iptables-persistent
sudo netfilter-persistent save

sudo systemctl enable docker
sudo systemctl start docker


#Створюємо файл підкачки на 2 Гігабайти
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

echo "Cloud-init успішно завершено! Docker готовий до роботи."


sudo ufw allow 80/tcp
sudo ufw allow 3000/tcp

sudo iptables -I INPUT 1 -p tcp --dport 80 -j ACCEPT
sudo iptables -I INPUT 1 -p tcp --dport 3000 -j ACCEPT
sudo netfilter-persistent save
sudo netfilter-persistent reload