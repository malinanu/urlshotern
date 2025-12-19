# Azure Deployment Guide

This guide covers deploying the URL Shortener application to Microsoft Azure with production-ready infrastructure.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Infrastructure Setup](#infrastructure-setup)
  - [Resource Group](#resource-group)
  - [Virtual Network](#virtual-network)
  - [Azure Database for PostgreSQL](#azure-database-for-postgresql)
  - [Azure Cache for Redis](#azure-cache-for-redis)
  - [App Service / Container Instances](#app-service--container-instances)
  - [Application Gateway](#application-gateway)
  - [Azure CDN](#azure-cdn)
- [Application Deployment](#application-deployment)
- [Secrets Management](#secrets-management)
- [Monitoring and Logging](#monitoring-and-logging)
- [Auto Scaling](#auto-scaling)
- [Backup and Disaster Recovery](#backup-and-disaster-recovery)
- [Cost Optimization](#cost-optimization)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                      Azure Front Door / CDN                   │
│                   (Global Edge Locations)                     │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│                  Application Gateway (WAF)                    │
│                      (SSL/TLS Termination)                    │
└─────────┬────────────────────────────────────┬───────────────┘
          │                                    │
┌─────────▼──────────┐              ┌─────────▼──────────┐
│  App Service       │              │  App Service       │
│  (Zone 1)          │              │  (Zone 2)          │
│  Auto-scale enabled│              │  Auto-scale enabled│
└─────────┬──────────┘              └─────────┬──────────┘
          │                                    │
          │      ┌──────────────────────┐      │
          └──────▶ Azure Cache for Redis◀──────┘
                 │  (Premium Tier)      │
                 └──────────────────────┘
                          │
                 ┌────────▼─────────┐
                 │ Azure Database   │
                 │ for PostgreSQL   │
                 │ (Flexible Server)│
                 └──────────────────┘
```

**Key Components:**
- **Azure Front Door**: Global CDN with intelligent routing
- **Application Gateway**: L7 load balancer with WAF
- **App Service**: Managed platform for containers
- **Azure Database for PostgreSQL**: Managed database service
- **Azure Cache for Redis**: Managed Redis service
- **Key Vault**: Secure secrets management
- **Application Insights**: Monitoring and telemetry
- **Azure Monitor**: Logging and metrics
- **Azure Storage**: Static assets and backups

---

## Prerequisites

### Required Tools
- Azure CLI 2.0+: `az --version`
- Docker: `docker --version`
- PostgreSQL client: `psql --version`

### Azure Account
- Active Azure subscription
- Azure CLI logged in: `az login`
- Contributor role on subscription

### Domain Setup
- Registered domain name
- Access to DNS management

---

## Infrastructure Setup

### Resource Group

```bash
# Create resource group
az group create \
  --name urlshortener-rg \
  --location eastus \
  --tags Environment=Production Application=URLShortener
```

---

### Virtual Network

```bash
# Create VNet
az network vnet create \
  --resource-group urlshortener-rg \
  --name urlshortener-vnet \
  --address-prefix 10.0.0.0/16 \
  --location eastus

# Create subnets
az network vnet subnet create \
  --resource-group urlshortener-rg \
  --vnet-name urlshortener-vnet \
  --name app-subnet \
  --address-prefix 10.0.1.0/24

az network vnet subnet create \
  --resource-group urlshortener-rg \
  --vnet-name urlshortener-vnet \
  --name db-subnet \
  --address-prefix 10.0.2.0/24

az network vnet subnet create \
  --resource-group urlshortener-rg \
  --vnet-name urlshortener-vnet \
  --name redis-subnet \
  --address-prefix 10.0.3.0/24

# Create Network Security Group
az network nsg create \
  --resource-group urlshortener-rg \
  --name app-nsg

az network nsg rule create \
  --resource-group urlshortener-rg \
  --nsg-name app-nsg \
  --name AllowHTTPS \
  --priority 100 \
  --source-address-prefixes Internet \
  --destination-port-ranges 443 \
  --access Allow \
  --protocol Tcp
```

---

### Azure Database for PostgreSQL

**Create Flexible Server:**

```bash
az postgres flexible-server create \
  --resource-group urlshortener-rg \
  --name urlshortener-db \
  --location eastus \
  --admin-user dbadmin \
  --admin-password '<SECURE_PASSWORD>' \
  --sku-name Standard_B2s \
  --tier Burstable \
  --storage-size 32 \
  --version 14 \
  --public-access None \
  --vnet urlshortener-vnet \
  --subnet db-subnet \
  --backup-retention 7 \
  --geo-redundant-backup Enabled \
  --high-availability Enabled \
  --zone 1 \
  --standby-zone 2 \
  --tags Environment=Production
```

**Pricing Tiers:**
| Tier | vCores | RAM | Storage | Cost/Month |
|------|--------|-----|---------|------------|
| Burstable (B2s) | 2 | 4GB | 32GB | ~$50 |
| General Purpose (D2s_v3) | 2 | 8GB | 128GB | ~$150 |
| Memory Optimized (E2s_v3) | 2 | 16GB | 256GB | ~$300 |

**Configure Firewall:**

```bash
# Allow Azure services
az postgres flexible-server firewall-rule create \
  --resource-group urlshortener-rg \
  --name urlshortener-db \
  --rule-name AllowAzureServices \
  --start-ip-address 0.0.0.0 \
  --end-ip-address 0.0.0.0
```

**Initialize Database:**

```bash
# Get connection string
az postgres flexible-server show-connection-string \
  --server-name urlshortener-db \
  --database-name postgres \
  --admin-user dbadmin

# Connect and create database
psql "host=urlshortener-db.postgres.database.azure.com port=5432 dbname=postgres user=dbadmin password=<PASSWORD> sslmode=require"

CREATE DATABASE urlshortener_prod;
\c urlshortener_prod

# Run migrations
psql "host=urlshortener-db.postgres.database.azure.com port=5432 dbname=urlshortener_prod user=dbadmin password=<PASSWORD> sslmode=require" < configs/database.sql
```

---

### Azure Cache for Redis

```bash
az redis create \
  --resource-group urlshortener-rg \
  --name urlshortener-redis \
  --location eastus \
  --sku Premium \
  --vm-size P1 \
  --enable-non-ssl-port false \
  --minimum-tls-version 1.2 \
  --redis-configuration maxmemory-policy=allkeys-lru \
  --zones 1 2 \
  --replicas-per-master 1 \
  --tags Environment=Production
```

**Pricing Tiers:**
| Tier | Size | Memory | Cost/Month |
|------|------|--------|------------|
| Basic | C1 | 1GB | ~$20 |
| Standard | C2 | 2.5GB | ~$80 |
| Premium | P1 | 6GB | ~$250 |

**Get Connection String:**

```bash
az redis show \
  --resource-group urlshortener-rg \
  --name urlshortener-redis \
  --query [hostName,sslPort,accessKeys.primaryKey] \
  --output tsv
```

---

### App Service

**1. Create Container Registry:**

```bash
az acr create \
  --resource-group urlshortener-rg \
  --name urlshorteneracr \
  --sku Standard \
  --admin-enabled true

# Get credentials
az acr credential show \
  --name urlshorteneracr \
  --query [username,passwords[0].value] \
  --output tsv
```

**2. Build and Push Docker Image:**

```bash
# Login to ACR
az acr login --name urlshorteneracr

# Build and push
docker build -t urlshorteneracr.azurecr.io/urlshortener:latest .
docker push urlshorteneracr.azurecr.io/urlshortener:latest
```

**3. Create App Service Plan:**

```bash
az appservice plan create \
  --resource-group urlshortener-rg \
  --name urlshortener-plan \
  --location eastus \
  --is-linux \
  --sku P1V2 \
  --number-of-workers 2 \
  --zone-redundant \
  --tags Environment=Production
```

**App Service Plan Pricing:**
| Tier | Cores | RAM | Instances | Cost/Month |
|------|-------|-----|-----------|------------|
| B1 | 1 | 1.75GB | 1 | ~$15 |
| P1V2 | 1 | 3.5GB | 2 | ~$140 |
| P2V2 | 2 | 7GB | 2 | ~$280 |

**4. Create Web Apps:**

```bash
# Primary instance (Zone 1)
az webapp create \
  --resource-group urlshortener-rg \
  --plan urlshortener-plan \
  --name urlshortener-app-z1 \
  --deployment-container-image-name urlshorteneracr.azurecr.io/urlshortener:latest

# Secondary instance (Zone 2)
az webapp create \
  --resource-group urlshortener-rg \
  --plan urlshortener-plan \
  --name urlshortener-app-z2 \
  --deployment-container-image-name urlshorteneracr.azurecr.io/urlshortener:latest
```

**5. Configure App Settings:**

```bash
az webapp config appsettings set \
  --resource-group urlshortener-rg \
  --name urlshortener-app-z1 \
  --settings \
    ENVIRONMENT=production \
    DB_HOST=urlshortener-db.postgres.database.azure.com \
    DB_PORT=5432 \
    DB_USER=dbadmin \
    DB_PASSWORD=@Microsoft.KeyVault(SecretUri=https://urlshortener-kv.vault.azure.net/secrets/db-password/) \
    REDIS_HOST=urlshortener-redis.redis.cache.windows.net \
    REDIS_PORT=6380 \
    REDIS_PASSWORD=@Microsoft.KeyVault(SecretUri=https://urlshortener-kv.vault.azure.net/secrets/redis-password/) \
    JWT_SECRET=@Microsoft.KeyVault(SecretUri=https://urlshortener-kv.vault.azure.net/secrets/jwt-secret/) \
    WEBSITES_PORT=8080
```

**6. Enable Managed Identity:**

```bash
az webapp identity assign \
  --resource-group urlshortener-rg \
  --name urlshortener-app-z1
```

---

### Application Gateway

```bash
# Create public IP
az network public-ip create \
  --resource-group urlshortener-rg \
  --name appgw-pip \
  --allocation-method Static \
  --sku Standard \
  --zone 1 2 3

# Create Application Gateway
az network application-gateway create \
  --resource-group urlshortener-rg \
  --name urlshortener-appgw \
  --location eastus \
  --sku WAF_v2 \
  --capacity 2 \
  --vnet-name urlshortener-vnet \
  --subnet app-subnet \
  --public-ip-address appgw-pip \
  --http-settings-cookie-based-affinity Disabled \
  --http-settings-protocol Https \
  --http-settings-port 443 \
  --frontend-port 443 \
  --servers urlshortener-app-z1.azurewebsites.net urlshortener-app-z2.azurewebsites.net
```

**Configure SSL Certificate:**

```bash
# Upload SSL certificate
az network application-gateway ssl-cert create \
  --resource-group urlshortener-rg \
  --gateway-name urlshortener-appgw \
  --name ssl-cert \
  --cert-file /path/to/cert.pfx \
  --cert-password '<CERT_PASSWORD>'
```

**Enable WAF:**

```bash
az network application-gateway waf-config set \
  --resource-group urlshortener-rg \
  --gateway-name urlshortener-appgw \
  --enabled true \
  --firewall-mode Prevention \
  --rule-set-type OWASP \
  --rule-set-version 3.2
```

---

### Azure CDN

```bash
# Create CDN profile
az cdn profile create \
  --resource-group urlshortener-rg \
  --name urlshortener-cdn \
  --sku Standard_Microsoft \
  --location Global

# Create CDN endpoint
az cdn endpoint create \
  --resource-group urlshortener-rg \
  --profile-name urlshortener-cdn \
  --name urlshortener-endpoint \
  --origin urlshortener-appgw-pip.eastus.cloudapp.azure.com \
  --origin-host-header yourdomain.com \
  --enable-compression true \
  --query-string-caching-behavior BypassCaching

# Add custom domain
az cdn custom-domain create \
  --resource-group urlshortener-rg \
  --endpoint-name urlshortener-endpoint \
  --profile-name urlshortener-cdn \
  --name yourdomain \
  --hostname yourdomain.com

# Enable HTTPS
az cdn custom-domain enable-https \
  --resource-group urlshortener-rg \
  --endpoint-name urlshortener-endpoint \
  --profile-name urlshortener-cdn \
  --name yourdomain
```

---

## Secrets Management

**1. Create Key Vault:**

```bash
az keyvault create \
  --resource-group urlshortener-rg \
  --name urlshortener-kv \
  --location eastus \
  --enable-rbac-authorization false \
  --enabled-for-deployment true \
  --enabled-for-template-deployment true
```

**2. Add Secrets:**

```bash
# Database password
az keyvault secret set \
  --vault-name urlshortener-kv \
  --name db-password \
  --value '<SECURE_DB_PASSWORD>'

# Redis password
az keyvault secret set \
  --vault-name urlshortener-kv \
  --name redis-password \
  --value '<SECURE_REDIS_PASSWORD>'

# JWT secret
az keyvault secret set \
  --vault-name urlshortener-kv \
  --name jwt-secret \
  --value '<SECURE_JWT_SECRET>'
```

**3. Grant App Service Access:**

```bash
# Get App Service managed identity principal ID
PRINCIPAL_ID=$(az webapp identity show \
  --resource-group urlshortener-rg \
  --name urlshortener-app-z1 \
  --query principalId \
  --output tsv)

# Grant access policy
az keyvault set-policy \
  --name urlshortener-kv \
  --object-id $PRINCIPAL_ID \
  --secret-permissions get list
```

---

## Monitoring and Logging

**1. Create Application Insights:**

```bash
az monitor app-insights component create \
  --resource-group urlshortener-rg \
  --app urlshortener-insights \
  --location eastus \
  --application-type web \
  --kind web

# Get instrumentation key
INSTRUMENTATION_KEY=$(az monitor app-insights component show \
  --resource-group urlshortener-rg \
  --app urlshortener-insights \
  --query instrumentationKey \
  --output tsv)
```

**2. Configure App Service:**

```bash
az webapp config appsettings set \
  --resource-group urlshortener-rg \
  --name urlshortener-app-z1 \
  --settings APPINSIGHTS_INSTRUMENTATIONKEY=$INSTRUMENTATION_KEY
```

**3. Create Alerts:**

```bash
# High CPU alert
az monitor metrics alert create \
  --name high-cpu \
  --resource-group urlshortener-rg \
  --scopes /subscriptions/<SUB_ID>/resourceGroups/urlshortener-rg/providers/Microsoft.Web/serverFarms/urlshortener-plan \
  --condition "avg Percentage CPU > 80" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --action /subscriptions/<SUB_ID>/resourceGroups/urlshortener-rg/providers/Microsoft.Insights/actionGroups/alerts

# Database connection alert
az monitor metrics alert create \
  --name high-db-connections \
  --resource-group urlshortener-rg \
  --scopes /subscriptions/<SUB_ID>/resourceGroups/urlshortener-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/urlshortener-db \
  --condition "avg active_connections > 80" \
  --window-size 5m \
  --evaluation-frequency 1m
```

---

## Auto Scaling

```bash
# Configure autoscale
az monitor autoscale create \
  --resource-group urlshortener-rg \
  --resource /subscriptions/<SUB_ID>/resourceGroups/urlshortener-rg/providers/Microsoft.Web/serverFarms/urlshortener-plan \
  --name urlshortener-autoscale \
  --min-count 2 \
  --max-count 10 \
  --count 2

# Scale out rule (CPU > 70%)
az monitor autoscale rule create \
  --resource-group urlshortener-rg \
  --autoscale-name urlshortener-autoscale \
  --condition "Percentage CPU > 70 avg 5m" \
  --scale out 1

# Scale in rule (CPU < 30%)
az monitor autoscale rule create \
  --resource-group urlshortener-rg \
  --autoscale-name urlshortener-autoscale \
  --condition "Percentage CPU < 30 avg 5m" \
  --scale in 1
```

---

## Backup and Disaster Recovery

### Database Backups

```bash
# Automated backups are enabled by default with 7-day retention

# Create manual backup
az postgres flexible-server backup create \
  --resource-group urlshortener-rg \
  --name urlshortener-db \
  --backup-name manual-backup-2025-12-19

# Restore from backup
az postgres flexible-server restore \
  --resource-group urlshortener-rg \
  --name urlshortener-db-restored \
  --source-server urlshortener-db \
  --restore-time "2025-12-19T03:00:00Z"
```

### Geo-Replication

```bash
# Enable geo-replication for Redis
az redis update \
  --resource-group urlshortener-rg \
  --name urlshortener-redis \
  --set replicasPerMaster=1
```

---

## Cost Optimization

**Estimated Monthly Costs (East US):**

| Component | Size | Monthly Cost |
|-----------|------|--------------|
| App Service (P1V2) | 2 workers | $140 |
| PostgreSQL (B2s) | 32GB | $50 |
| Redis (P1) | 6GB | $250 |
| Application Gateway (WAF_v2) | Small | $250 |
| CDN (Standard) | 1TB | $80 |
| Application Insights | 5GB/day | $30 |
| **Total** | | **~$800/month** |

**Cost Reduction:**
1. Use Reserved Instances (40-60% savings)
2. Enable auto-shutdown for dev/test environments
3. Use Azure Hybrid Benefit if you have licenses
4. Implement aggressive CDN caching
5. Use Azure Cost Management for monitoring
6. Consider Spot VMs for non-critical workloads

---

## Troubleshooting

### View Application Logs

```bash
# Stream logs
az webapp log tail \
  --resource-group urlshortener-rg \
  --name urlshortener-app-z1

# Download logs
az webapp log download \
  --resource-group urlshortener-rg \
  --name urlshortener-app-z1 \
  --log-file logs.zip
```

### Check Database Connectivity

```bash
# Test connection
psql "host=urlshortener-db.postgres.database.azure.com port=5432 dbname=urlshortener_prod user=dbadmin password=<PASSWORD> sslmode=require" -c "SELECT 1"
```

### Redis Connectivity

```bash
# Test from App Service console
redis-cli -h urlshortener-redis.redis.cache.windows.net -p 6380 -a <PASSWORD> --tls ping
```

For more troubleshooting, see [TROUBLESHOOTING.md](../TROUBLESHOOTING.md).

---

**Last Updated:** 2025-12-19
**Maintained By:** DevOps Team
