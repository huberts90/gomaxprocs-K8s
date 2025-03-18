#!/bin/bash
set -e

kind delete cluster

# Limit the CPU to the value defined in [node_CPU]. Since Kind does not support the direct way to
# set this limit, the workaround is to reserve the rest of CPU for the K8s daemons like the kubelet, container runtime, etc.
node_CPU=4
kube_reserved=$(($(sysctl -n hw.ncpu)-node_CPU))
# Create kind cluster
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30080
    hostPort: 30080
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        cpu-manager-policy: "static"
        kube-reserved: "cpu=$kube_reserved"
EOF

# Build the Docker image
docker build -t gomaxprocs-k8s:1.0 .

# Load the image into kind cluster
kind load docker-image gomaxprocs-k8s:1.0

# Apply the Kubernetes manifests
kubectl apply -f k8s-deployment.yaml

# Wait for the deployment to be ready
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/gomaxprocs-k8s

echo "Setup complete! You can access the application at http://localhost:30080"
echo "To test the application, run: curl http://localhost:30080" 