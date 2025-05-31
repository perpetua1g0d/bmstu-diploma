docker pull docker.io/grafana/grafana:9.5.3
docker save -o grafana-9.5.3.tar docker.io/grafana/grafana:9.5.3

# deploy.sh:
# ...
docker load -i grafana-9.5.3.tar
k3d image import docker.io/grafana/grafana:9.5.3 -c bmstucluster
# ...
