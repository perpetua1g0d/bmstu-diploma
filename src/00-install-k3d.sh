#!/bin/bash
sudo apt-get update
sudo apt-get install -y docker.io curl
sudo usermod -aG docker $USER
newgrp docker

curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d cluster create pg-cluster --servers 1 --agents 2
