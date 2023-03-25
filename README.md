# Custom Kuberhealthy Synthetic Checks

Collection of [kuberhealthy](https://github.com/kuberhealthy/kuberhealthy) custom checks.

## Usage

- Deploy [kuberhealthy](https://github.com/kuberhealthy/kuberhealthy) to you cluster.
- See `check.yaml` in `*-check` directories for examples. Adjust these with custom parameters and deploy (`k apply -f ...`) them to Kubernetes cluster.

Example:

```shell
# Create cluster:
minikube delete && minikube start \
  --kubernetes-version=v1.26.1 \
  --memory=6g \
  --bootstrapper=kubeadm \
  --extra-config=kubelet.authentication-token-webhook=true \
  --extra-config=kubelet.authorization-mode=Webhook \
  --extra-config=scheduler.bind-address=0.0.0.0 \
  --extra-config=controller-manager.bind-address=0.0.0.0
  
# Deploy kuberhealhy:
helm repo add kuberhealthy https://kuberhealthy.github.io/kuberhealthy/helm-repos
helm install -n kuberhealthy kuberhealthy kuberhealthy/kuberhealthy --create-namespace  # --values values.yaml

# Deploy check
kubectl apply -f jq-check/check.yaml
# ... check logs of check Pod(s)

kubectl port-forward -n kuberhealthy svc/kuberhealthy 8080:80
# ... check "localhost:8080" and "localhost:8080/metrics" for check reports
```
