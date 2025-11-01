# simple-azure-aks-app-go

A simple Hello World Go application with structured logging using Uber's Zap package, containerized with Docker and ready for deployment to Azure Kubernetes Service (AKS).

## Features

- **Simple HTTP Server**: Responds with "Hello, World!" message on the root endpoint
- **Structured Logging**: Uses `go.uber.org/zap` for production-grade structured logging in JSON format
- **Health Check Endpoint**: `/health` endpoint for Kubernetes liveness and readiness probes
- **Request Logging**: Logs request start and completion with method, path, duration, and status
- **Containerized**: Multi-stage Docker build for minimal image size using distroless base image
- **Kubernetes Ready**: Includes deployment and service manifests for Azure AKS

## Endpoints

- `GET /` - Returns hello world message
- `GET /health` - Health check endpoint (returns OK)

## Local Development

### Prerequisites

- Go 1.23 or later
- Docker (for containerization)

### Build and Run Locally

```bash
# Build the application
go build -o main .

# Run the application
./main

# Test the endpoints
curl http://localhost:8080/
curl http://localhost:8080/health
```

The application will start on port 8080 by default. You can override this by setting the `PORT` environment variable.

## Docker

### Build Docker Image

```bash
docker build -t hello-world-go:latest .
```

### Run Docker Container

```bash
docker run -p 8080:8080 hello-world-go:latest
```

## Kubernetes Deployment to Azure AKS

### Prerequisites

- Azure CLI installed and configured
- kubectl installed
- An Azure Container Registry (ACR)
- An Azure Kubernetes Service (AKS) cluster

### Deploy to AKS

1. **Build and push the Docker image to ACR:**

```bash
# Login to Azure
az login

# Create a resource group (if needed)
az group create --name myResourceGroup --location eastus

# Create an ACR (if needed)
az acr create --resource-group myResourceGroup --name <ACR_NAME> --sku Basic

# Login to ACR
az acr login --name <ACR_NAME>

# Tag and push the image
docker tag hello-world-go:latest <ACR_NAME>.azurecr.io/hello-world-go:latest
docker push <ACR_NAME>.azurecr.io/hello-world-go:latest
```

2. **Create an AKS cluster (if needed):**

```bash
az aks create \
  --resource-group myResourceGroup \
  --name myAKSCluster \
  --node-count 2 \
  --enable-managed-identity \
  --attach-acr <ACR_NAME> \
  --generate-ssh-keys
```

3. **Connect to the AKS cluster:**

```bash
az aks get-credentials --resource-group myResourceGroup --name myAKSCluster
```

4. **Update the deployment manifest:**

Edit `k8s/deployment.yaml` and replace `<ACR_NAME>` with your actual ACR name.

5. **Deploy the application:**

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

6. **Get the external IP:**

```bash
kubectl get service hello-world-go
```

Wait for the `EXTERNAL-IP` to be assigned, then access your application at `http://<EXTERNAL-IP>/`

## Structured Logging

The application uses Zap for structured logging. All logs are output in JSON format for easy parsing and integration with log aggregation systems like Azure Monitor, ELK stack, or similar.

Example log output:
```json
{"level":"info","ts":1761960955.3738174,"caller":"app/main.go:33","msg":"Starting server","port":"8080"}
{"level":"info","ts":1761960972.3560178,"caller":"app/main.go:48","msg":"Request started","method":"GET","path":"/","remote_addr":"172.17.0.1:45632"}
{"level":"info","ts":1761960972.3560603,"caller":"app/main.go:59","msg":"Request completed","method":"GET","path":"/","duration":0.000046917,"status":200}
```

## Monitoring in Azure

The structured logs can be viewed in Azure Monitor:

1. Enable Container Insights for your AKS cluster
2. View logs in Azure Log Analytics workspace
3. Query logs using Kusto Query Language (KQL)

Example query:
```kql
ContainerLog
| where ContainerName == "hello-world-go"
| project TimeGenerated, LogEntry
| order by TimeGenerated desc
```

