# Azure Blob Storage Setup for Fluent Bit

## 1. Create Azure Storage Account

```bash
# Set variables
RESOURCE_GROUP="github-logging"
STORAGE_ACCOUNT="01h10sgithublogs"
LOCATION="centralus"

# Create storage account
az storage account create \
  --name $STORAGE_ACCOUNT \
  --resource-group $RESOURCE_GROUP \
  --location $LOCATION \
  --sku Standard_LRS \
  --kind StorageV2
```

## 2. Get Storage Account Key

```bash
# Get the storage account key
STORAGE_KEY=$(az storage account keys list \
  --resource-group $RESOURCE_GROUP \
  --account-name $STORAGE_ACCOUNT \
  --query '[0].value' \
  --output tsv)

echo "Storage Key: $STORAGE_KEY"
```

## 3. Create blob container

```bash
az storage container create \
  --name logs \
  --account-name $STORAGE_ACCOUNT \
  --account-key $STORAGE_KEY
```

## 4. Create Kubernetes Secret

```bash
# Create or update the secret with storage key
kubectl create secret generic fluent-bit-azure-blob-credentials \
  --namespace kube-system \
  --from-literal=storage-key="$STORAGE_KEY" \
  --dry-run=client -o yaml | kubectl apply -f -
```

## 5. Update ConfigMap

Update the `fluent-bit-azure-blob-config.yaml` file with your actual storage account name:

```yaml
data:
  storage-account: "01h10sgithublogs"
  blob-container: "logs"
```

## 6. Apply Configuration

```bash
kubectl apply -f k8s/fluent-bit-azure-blob-config.yaml
kubectl apply -f k8s/fluent-bit-configmap.yaml
kubectl rollout restart daemonset/fluent-bit -n kube-system
```
