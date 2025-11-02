## Steps

### Step 1. Create a custom table

This table must incude a `TimeGenerated` column.

```
az monitor log-analytics workspace table create \
  --resource-group github-logging \
  --workspace-name azure-monitor-logs-test-1 \
  --name MyFluentBitLogs_CL \
  --columns TimeGenerated=datetime level=string caller=string msg=string method=string path=string data=dynamic \
  --retention-time 30
```

https://learn.microsoft.com/en-us/cli/azure/monitor/log-analytics/workspace/table?view=azure-cli-latest#az-monitor-log-analytics-workspace-table-create

### Step 2. Create a data collection rule

```
az monitor data-collection rule create \
  --resource-group github-logging \
  --name dcr-for-fluentbit \
  --location centralus \
  --rule-file dcr.json
```

https://learn.microsoft.com/en-us/cli/azure/monitor/data-collection/rule?view=azure-cli-latest#az-monitor-data-collection-rule-create
