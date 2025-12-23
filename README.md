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
docker build --platform linux/amd64 -t hello-world-go:latest .
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

# Build with a ACR tag
docker build --platform linux/amd64 -t <ACR_NAME>.azurecr.io/hello-world-go:latest .

# Push the image
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

## Fluent Bit Log Collection

This repository includes Fluent Bit DaemonSet manifests for collecting and forwarding logs to Azure Monitor using the `azure_logs_ingestion` output plugin.

### Prerequisites for Fluent Bit

1. **Azure Log Analytics Workspace**: You need a Log Analytics workspace to receive the logs
2. **Data Collection Endpoint (DCE)**: Create a DCE in your Azure subscription
3. **Data Collection Rule (DCR)**: Create a DCR that defines how logs are ingested
4. **Azure Service Principal**: Create a service principal with permissions to send logs
   - Required permissions: `Monitoring Metrics Publisher` role on the DCR

### Setting up Azure Resources

1. **Create a Log Analytics workspace:**

```bash
az monitor log-analytics workspace create \
  --resource-group myResourceGroup \
  --workspace-name myLogAnalyticsWorkspace \
  --location eastus
```

2. **Create a Data Collection Endpoint:**

```bash
az monitor data-collection endpoint create \
  --name myDCE \
  --resource-group myResourceGroup \
  --location eastus \
  --public-network-access Enabled
```

3. **Create Custom Tables:**

Tables must incude a `TimeGenerated` column.

#### Primary custom table

```
az monitor log-analytics workspace table create \
  --resource-group github-logging \
  --workspace-name azure-monitor-logs-test-1 \
  --name MyFluentBitLogs_CL \
  --columns TimeGenerated=datetime level=string caller=string msg=string method=string path=string data=dynamic \
  --total-retention-time 730
```

#### Short retention time

For testing long-term retention and search jobs a few days after population

```
az monitor log-analytics workspace table create \
  --resource-group github-logging \
  --workspace-name azure-monitor-logs-test-1 \
  --name ShortRetentionTime_CL \
  --columns TimeGenerated=datetime level=string caller=string msg=string method=string path=string data=dynamic \
  --retention-time 4 \
  --total-retention-time 730
```

#### A table with artifically old TimeGenerated values

For immediately testing long-term retention and search jobs after population

```
az monitor log-analytics workspace table create \
  --resource-group github-logging \
  --workspace-name azure-monitor-logs-test-1 \
  --name TimeShifted_CL \
  --columns TimeGenerated=datetime level=string caller=string msg=string method=string path=string data=dynamic \
  --total-retention-time 730
```

https://learn.microsoft.com/en-us/cli/azure/monitor/log-analytics/workspace/table?view=azure-cli-latest#az-monitor-log-analytics-workspace-table-create

4. **Create a Data Collection Rule:**

```
az monitor data-collection rule create \
  --resource-group github-logging \
  --name dcr-for-fluentbit \
  --location centralus \
  --rule-file custom-schema/dcr.json
```

https://learn.microsoft.com/en-us/azure/azure-monitor/data-collection/data-collection-rule-samples?source=recommendations#logs-ingestion-api

https://learn.microsoft.com/en-us/azure/azure-monitor/data-collection/data-collection-rule-structure

https://learn.microsoft.com/en-us/cli/azure/monitor/data-collection/rule?view=azure-cli-latest#az-monitor-data-collection-rule-create

5. **Configure the DCE for the new rule:**

Find the new Data Collection Rule in the Azure Portal and click "Configure DCE" in the Overview section. Select the DCE created in step 2.

6. **Create a Service Principal:**

```bash
az ad sp create-for-rbac --name fluent-bit-logger --role "Monitoring Metrics Publisher" --scopes /subscriptions/<SUBSCRIPTION_ID>/resourceGroups/myResourceGroup
```

Note the `appId` (client ID), `password` (client secret), and `tenant` (tenant ID) from the output.

### Deploying Fluent Bit to AKS

1. **Edit the Azure configuration:**

Edit `k8s/fluent-bit-azure-config.yaml` and update the following values:
- `dce-url`: Your Data Collection Endpoint URL (e.g., `https://myDCE.eastus.ingest.monitor.azure.com`)
- `dcr-id`: Your Data Collection Rule immutable ID (e.g., `dcr-abc123def456`)
- `table-name`: The custom table name in Log Analytics (e.g., `CustomLog_CL`)

2. **Apply Fluent Bit config:**

```bash
# Apply RBAC permissions
kubectl apply -f k8s/fluent-bit-rbac.yaml

# Apply configuration
kubectl apply -f k8s/fluent-bit-configmap.yaml
kubectl apply -f k8s/fluent-bit-azure-config.yaml
```

3. **Create the Azure credentials secret:**

Override any secret possibly created by `k8s/fluent-bit-azure-config.yaml`.

```bash
kubectl create secret generic fluent-bit-azure-credentials \
  --from-literal=client-id=<YOUR_CLIENT_ID> \
  --from-literal=client-secret=<YOUR_CLIENT_SECRET> \
  --from-literal=tenant-id=<YOUR_TENANT_ID> \
  -n kube-system
```

4. **Deploy Fluent Bit:**

```bash
# Deploy the DaemonSet
kubectl apply -f k8s/fluent-bit-daemonset.yaml
```

5. **Verify Fluent Bit is running:**

```bash
kubectl get daemonset fluent-bit -n kube-system
kubectl get pods -n kube-system -l app=fluent-bit
kubectl logs -n kube-system -l app=fluent-bit --tail=50
```

### Viewing Logs in Azure

Once Fluent Bit is running and configured, logs will be sent to your Log Analytics workspace. Query them using KQL:

```kql
CustomLog_CL
| where TimeGenerated > ago(1h)
| project TimeGenerated, Log, kubernetes_pod_name, kubernetes_namespace_name
| order by TimeGenerated desc
```

### Fluent Bit Configuration

The Fluent Bit configuration includes:

- **Inputs**:
  - `tail`: Reads container logs from `/var/log/containers/*.log`
  - `systemd`: Reads kubelet service logs

- **Filters**:
  - `kubernetes`: Enriches logs with Kubernetes metadata (pod name, namespace, labels, etc.)

- **Output**:
  - `azure_logs_ingestion`: Sends logs to Azure Monitor using the Logs Ingestion API

**Container Runtime Compatibility:**
The DaemonSet is configured to work with both Docker and containerd container runtimes. It mounts `/var/log` and `/var/lib/docker/containers` to access container logs. Modern AKS clusters typically use containerd, which stores logs in `/var/log/pods` and creates symlinks in `/var/log/containers`.

The configuration files are located in `k8s/`:
- `fluent-bit-configmap.yaml`: Fluent Bit configuration and parsers
- `fluent-bit-daemonset.yaml`: DaemonSet specification
- `fluent-bit-rbac.yaml`: ServiceAccount and RBAC permissions
- `fluent-bit-azure-config.yaml`: Azure-specific configuration (DCE, DCR, table name)
