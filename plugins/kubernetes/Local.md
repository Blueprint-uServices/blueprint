# Locally Testing Kubernetes

## Running a Docker Registry

The following command runs a local docker registry

```bash
docker run -d --restart=always --name registry -p 5000:5000 registry:2
```

## Setting Up a Local Cluster

### Installing Kind

```
curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Installing Kubectl

```
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod a+x kubectl
sudo mv kubectl /usr/local/bin
```

### Creating a Cluster

```
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
    endpoint = ["http://registry:5000"]
EOF
```

### Connecting Cluster to Registry

```
docker network connect kind registry
```