# JsonServer Kubernetes Operator (GitOps-enabled)

This repository contains a **full Kubernetes Operator implementation** for managing `json-server` instances using **Kubebuilder**, **CustomResourceDefinitions (CRDs)**, **Validating Admission Webhooks**, and **Controllers**, deployed via **GitOps with Flux** and **CI-driven image updates**.

This README is a **complete, step-by-step guide**, starting from an empty machine to a fully working system, including **Flux bootstrap with Git authentication**, **component installation**, and **GitOps reconciliation**.

---

## 1. What This Project Does

The operator introduces a new Kubernetes resource:

```yaml
apiVersion: example.com/v1
kind: JsonServer
```

When a `JsonServer` resource is created, the controller automatically:

- Creates a `ConfigMap` containing `db.json`
- Creates a `Deployment running backplane/json-server`
- Creates a `Service` exposing port `3000`
- Reconciles changes on update
- Deletes all child resources on delete
- Updates `.status` with `Synced` or `Error`

A **validating webhook** enforces:
- `metadata.name` must start with `app-`
- `spec.jsonConfig` must be valid JSON

---

## 2. Supported Platforms

| Platform | Supported | Notes |
|--------|-----------|------|
| macOS (Intel / Apple Silicon) | ✅ | Recommended |
| Linux | ✅ | Not Tested |
| Windows (WSL2) | ✅ | Not Tested |

All commands are identical across platforms (run inside WSL on Windows).

---

## 3. Prerequisites

Install the following tools:

- Docker
- kubectl
- kind
- Go >= 1.22
- kubebuilder
- Flux CLI

### macOS
```bash
brew install docker kubectl kind go fluxcd/tap/flux
```

### Linux (Ubuntu)
```bash
sudo apt update
sudo apt install -y docker.io kubectl golang
curl -s https://fluxcd.io/install.sh | sudo bash
```

### Windows (WSL2)
```powershell
wsl --install
```
Install Linux tools inside WSL.

---

## 4. Create a Local Kubernetes Cluster

```bash
kind create cluster --name json-server
kubectl cluster-info
```

---

## 5. Install cert-manager (Required for Webhooks)

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
kubectl wait --for=condition=Available deployment/cert-manager -n cert-manager --timeout=120s
```

---

## 6. Build and Deploy the Operator

### Generate CRDs and RBAC
```bash
make manifests
```

### Run Controller Locally (Dev Mode)
```bash
make run
```

### OR Deploy Controller to the Cluster
```bash
make docker-build docker-push IMG=controller:latest
kind load docker-image controller:latest
make deploy IMG=controller:latest
```

Verify:
```bash
kubectl get pods -n json-server-system
```

---

## 7. Bootstrap Flux with Git Authentication

Flux must be bootstrapped with **Git user identity and authentication** so it can:
- Install Flux components
- Push changes back to Git (if needed)
- Reconcile GitOps manifests

### 7.1 Create a GitHub Personal Access Token (PAT)

Create a token with:
- Repository access (read/write)
- Contents: Read & Write

Export credentials locally:

```bash
export GITHUB_USER=<your-github-username>
export GITHUB_TOKEN=<your-github-token>
```

---

### 7.2 Bootstrap Flux (with extra components)

```bash
flux bootstrap github   
  --owner=$GITHUB_USER  
  --repository=json-server-flux  
  --branch=main   
  --path=clusters/json-server   
  --personal   
  --components-extra=image-reflector-controller,image-automation-controller   
  --token-auth
```

This command:
- Creates the `flux-system` namespace
- Installs Flux controllers
- Installs extra image components
- Commits bootstrap manifests to Git
- Configures Git authentication

Verify:
```bash
flux get all
```

---

## 8. Apply GitOps Manifests (Application Repo)

The `config/gitops/` directory contains:

- `GitRepository` – points Flux to this repository
- `Kustomization` – applies `config/default`

Apply them:

```bash
kubectl apply -f config/gitops/
```

Verify:
```bash
flux get sources git
flux get kustomizations
```

---

## 9. CI and Image Management (ttl.sh)

- GitHub Actions builds the controller image
- Pushes it to `ttl.sh`
- Updates `config/manager/kustomization.yaml`
- Commits the change back to Git
- Flux reconciles automatically

> `ttl.sh` requires the TTL to be encoded in the image tag (e.g. `:2h`), so CI-driven GitOps updates are used instead of Flux image automation.

---
## 10. Testing

Test samples files are be found /config/sample

### 10.1 Deploy a VALID JsonServer

```yaml
apiVersion: example.com/v1
kind: JsonServer
metadata:
  name: app-basic
spec:
  replicas: 1
  jsonConfig: |
    {
      "people": [
        { "id": 1, "name": "Alice" }
      ]
    }
```

```bash
kubectl apply -f valid.yaml
```

Verify created resources:
```bash
kubectl get deploy,svc,cm
```

---

### 10.2 Test the Running json-server

```bash
kubectl port-forward svc/app-basic 3000:3000
curl http://localhost:3000/people
```

Expected:
```json
[{ "id": 1, "name": "Alice" }]
```

---

### 10.3 INVALID JsonServer Tests

### Invalid Name
```yaml
metadata:
  name: my-server
```
Result:
```
metadata.name must start with app-
```

### Invalid JSON
```yaml
jsonConfig: |
  { invalid json
```
Result:
```
Error: spec.jsonConfig is not a valid json object
```

---

## 10.4. Scaling with kubectl

```bash
kubectl scale jsonserver app-basic --replicas=3
kubectl get deploy app-basic
```

---

## 10.5 Status Reporting

```bash
kubectl get jsonserver app-basic -o yaml
```

Example:
```yaml
status:
  state: Synced
  message: Synced successfully!
  replicas: 3
```

---

## 11. Cleanup

```bash
make undeploy
kind delete cluster --name json-server
```

---

## 12. Design Notes

- Validating webhook blocks invalid resources early
- Controller updates status for runtime failures
- CI-driven GitOps used due to ttl.sh limitations
- Flux is bootstrapped with authenticated Git access
- Git remains the single source of truth

---

## 13. Summary

This project demonstrates:

- Kubernetes API extension with CRDs
- Controller reconciliation patterns
- Admission webhooks
- Status subresources
- GitOps with Flux
- CI-driven deployment
- Cross-platform local development
- Real-world GitOps trade-offs

