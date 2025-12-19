# Google Cloud Platform Deployment Guide

This guide covers deploying the URL Shortener application to Google Cloud Platform (GCP) with production-ready infrastructure.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Infrastructure Setup](#infrastructure-setup)
  - [Project and VPC](#project-and-vpc)
  - [Cloud SQL PostgreSQL](#cloud-sql-postgresql)
  - [Memorystore Redis](#memorystore-redis)
  - [Google Kubernetes Engine (GKE)](#google-kubernetes-engine-gke)
  - [Cloud Load Balancing](#cloud-load-balancing)
  - [Cloud CDN](#cloud-cdn)
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
│                      Cloud CDN                               │
│                  (Global Edge Network)                       │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│              Global HTTP(S) Load Balancer                    │
│                  (Cloud Armor WAF)                           │
└─────────┬──────────────────────────────────┬────────────────┘
          │                                  │
┌─────────▼──────────┐              ┌───────▼────────────┐
│  GKE Cluster       │              │  GKE Cluster       │
│  (Zone A)          │              │  (Zone B)          │
│  ┌──────────┐      │              │  ┌──────────┐      │
│  │ Pod      │      │              │  │ Pod      │      │
│  │ Auto-scale│     │              │  │ Auto-scale│     │
│  └──────────┘      │              │  └──────────┘      │
└──────┬─────────────┘              └──────┬─────────────┘
       │                                   │
       │      ┌──────────────────────┐     │
       └──────▶ Memorystore Redis    ◀─────┘
              │ (High Availability)  │
              └──────────────────────┘
                       │
              ┌────────▼──────────┐
              │  Cloud SQL        │
              │  PostgreSQL       │
              │  (HA Configuration)│
              └───────────────────┘
```

**Key Components:**
- **Cloud CDN**: Global content delivery network
- **Cloud Load Balancing**: Global L7 load balancer with SSL
- **GKE**: Managed Kubernetes service
- **Cloud SQL**: Managed PostgreSQL database
- **Memorystore**: Managed Redis service
- **Secret Manager**: Secure secrets storage
- **Cloud Monitoring**: Metrics and logging
- **Cloud Storage**: Static assets and backups
- **Cloud Armor**: DDoS protection and WAF

---

## Prerequisites

### Required Tools
- gcloud CLI: `gcloud version`
- kubectl: `kubectl version`
- Docker: `docker --version`
- Terraform (optional): `terraform --version`

### GCP Account
- GCP project with billing enabled
- gcloud CLI configured: `gcloud init`
- Required APIs enabled:
  - Compute Engine API
  - Kubernetes Engine API
  - Cloud SQL Admin API
  - Redis API
  - Secret Manager API

### Enable APIs

```bash
gcloud services enable compute.googleapis.com
gcloud services enable container.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable redis.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
```

---

## Infrastructure Setup

### Project and VPC

**1. Set Project:**

```bash
export PROJECT_ID=urlshortener-prod
export REGION=us-central1
export ZONE=us-central1-a

gcloud config set project $PROJECT_ID
gcloud config set compute/region $REGION
gcloud config set compute/zone $ZONE
```

**2. Create VPC:**

```bash
# Create VPC network
gcloud compute networks create urlshortener-vpc \
  --subnet-mode=custom \
  --bgp-routing-mode=regional

# Create subnets
gcloud compute networks subnets create gke-subnet \
  --network=urlshortener-vpc \
  --region=$REGION \
  --range=10.0.0.0/24 \
  --secondary-range pods=10.4.0.0/14,services=10.0.4.0/19

gcloud compute networks subnets create db-subnet \
  --network=urlshortener-vpc \
  --region=$REGION \
  --range=10.0.8.0/24

# Create firewall rules
gcloud compute firewall-rules create allow-internal \
  --network=urlshortener-vpc \
  --allow=tcp,udp,icmp \
  --source-ranges=10.0.0.0/8

gcloud compute firewall-rules create allow-health-checks \
  --network=urlshortener-vpc \
  --allow=tcp \
  --source-ranges=130.211.0.0/22,35.191.0.0/16
```

---

### Cloud SQL PostgreSQL

**1. Create SQL Instance:**

```bash
gcloud sql instances create urlshortener-db \
  --database-version=POSTGRES_14 \
  --tier=db-custom-2-7680 \
  --region=$REGION \
  --network=projects/$PROJECT_ID/global/networks/urlshortener-vpc \
  --no-assign-ip \
  --availability-type=REGIONAL \
  --backup-start-time=03:00 \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=4 \
  --enable-bin-log \
  --storage-auto-increase \
  --storage-size=20GB \
  --storage-type=SSD \
  --database-flags=max_connections=200
```

**Pricing Tiers:**
| Tier | vCPUs | RAM | Storage | Cost/Month |
|------|-------|-----|---------|------------|
| db-f1-micro | 1 shared | 0.6GB | 10GB | ~$10 |
| db-custom-1-3840 | 1 | 3.75GB | 20GB | ~$50 |
| db-custom-2-7680 | 2 | 7.5GB | 50GB | ~$100 |
| db-custom-4-15360 | 4 | 15GB | 100GB | ~$200 |

**2. Set Root Password:**

```bash
gcloud sql users set-password postgres \
  --instance=urlshortener-db \
  --password='<SECURE_PASSWORD>'
```

**3. Create Database and User:**

```bash
# Create database
gcloud sql databases create urlshortener_prod \
  --instance=urlshortener-db

# Create application user
gcloud sql users create appuser \
  --instance=urlshortener-db \
  --password='<SECURE_APP_PASSWORD>'

# Grant privileges (connect via Cloud SQL Proxy)
gcloud sql connect urlshortener-db --user=postgres

GRANT ALL PRIVILEGES ON DATABASE urlshortener_prod TO appuser;
\c urlshortener_prod
GRANT ALL PRIVILEGES ON SCHEMA public TO appuser;
```

**4. Run Migrations:**

```bash
# Download Cloud SQL Proxy
wget https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy
chmod +x cloud_sql_proxy

# Get connection name
INSTANCE_CONNECTION_NAME=$(gcloud sql instances describe urlshortener-db \
  --format='value(connectionName)')

# Start proxy
./cloud_sql_proxy -instances=$INSTANCE_CONNECTION_NAME=tcp:5432 &

# Run migrations
psql "host=127.0.0.1 port=5432 dbname=urlshortener_prod user=appuser password=<PASSWORD>" < configs/database.sql
```

---

### Memorystore Redis

```bash
gcloud redis instances create urlshortener-redis \
  --size=1 \
  --region=$REGION \
  --zone=$ZONE \
  --alternative-zone=${REGION}-b \
  --redis-version=redis_7_0 \
  --network=urlshortener-vpc \
  --tier=standard-ha \
  --enable-auth \
  --persistence-mode=rdb \
  --rdb-snapshot-period=12h \
  --display-name="URL Shortener Redis"
```

**Pricing Tiers:**
| Tier | Size | Type | Cost/Month |
|------|------|------|------------|
| Basic | 1GB | Single node | ~$30 |
| Standard | 1GB | HA (2 nodes) | ~$80 |
| Standard | 5GB | HA (2 nodes) | ~$250 |

**Get Connection Info:**

```bash
# Get host and port
gcloud redis instances describe urlshortener-redis \
  --region=$REGION \
  --format='value(host,port,authString)'
```

---

### Google Kubernetes Engine (GKE)

**1. Create GKE Cluster:**

```bash
gcloud container clusters create urlshortener-cluster \
  --region=$REGION \
  --network=urlshortener-vpc \
  --subnetwork=gke-subnet \
  --cluster-secondary-range-name=pods \
  --services-secondary-range-name=services \
  --enable-ip-alias \
  --enable-autoscaling \
  --min-nodes=1 \
  --max-nodes=10 \
  --num-nodes=2 \
  --machine-type=e2-medium \
  --disk-size=50GB \
  --enable-autorepair \
  --enable-autoupgrade \
  --enable-stackdriver-kubernetes \
  --addons=HorizontalPodAutoscaling,HttpLoadBalancing,GcePersistentDiskCsiDriver \
  --workload-pool=$PROJECT_ID.svc.id.goog \
  --enable-shielded-nodes \
  --release-channel=regular
```

**Node Pool Sizing:**
| Machine Type | vCPUs | RAM | Disk | Cost/Node/Month |
|--------------|-------|-----|------|-----------------|
| e2-micro | 2 shared | 1GB | 10GB | ~$8 |
| e2-small | 2 shared | 2GB | 20GB | ~$15 |
| e2-medium | 2 | 4GB | 50GB | ~$30 |
| n2-standard-2 | 2 | 8GB | 50GB | ~$75 |

**2. Get Credentials:**

```bash
gcloud container clusters get-credentials urlshortener-cluster \
  --region=$REGION
```

**3. Create Namespace:**

```bash
kubectl create namespace production
kubectl config set-context --current --namespace=production
```

---

### Container Registry

```bash
# Enable Container Registry
gcloud services enable containerregistry.googleapis.com

# Configure Docker
gcloud auth configure-docker

# Build and push image
docker build -t gcr.io/$PROJECT_ID/urlshortener:latest .
docker push gcr.io/$PROJECT_ID/urlshortener:latest
```

---

## Application Deployment

**1. Create Kubernetes Secrets:**

```bash
# Database credentials
kubectl create secret generic db-credentials \
  --from-literal=password='<DB_PASSWORD>'

# Redis credentials
kubectl create secret generic redis-credentials \
  --from-literal=auth-string='<REDIS_AUTH_STRING>'

# JWT secret
kubectl create secret generic jwt-secret \
  --from-literal=secret='<JWT_SECRET>'
```

**2. Deploy Application:**

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: urlshortener
  namespace: production
spec:
  replicas: 2
  selector:
    matchLabels:
      app: urlshortener
  template:
    metadata:
      labels:
        app: urlshortener
    spec:
      serviceAccountName: urlshortener-sa
      containers:
      - name: app
        image: gcr.io/PROJECT_ID/urlshortener:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: DB_HOST
          value: "127.0.0.1"  # Cloud SQL Proxy
        - name: DB_PORT
          value: "5432"
        - name: DB_USER
          value: "appuser"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        - name: DB_NAME
          value: "urlshortener_prod"
        - name: REDIS_HOST
          value: "REDIS_IP"
        - name: REDIS_PORT
          value: "6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-credentials
              key: auth-string
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

      # Cloud SQL Proxy sidecar
      - name: cloud-sql-proxy
        image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.1.0
        args:
          - "--structured-logs"
          - "--port=5432"
          - "PROJECT_ID:REGION:urlshortener-db"
        securityContext:
          runAsNonRoot: true
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: urlshortener
  namespace: production
spec:
  type: ClusterIP
  selector:
    app: urlshortener
  ports:
  - port: 80
    targetPort: 8080
```

**3. Apply Deployment:**

```bash
kubectl apply -f k8s/deployment.yaml
```

**4. Create Ingress with Cloud Load Balancer:**

Create `k8s/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: urlshortener-ingress
  namespace: production
  annotations:
    kubernetes.io/ingress.class: "gce"
    kubernetes.io/ingress.global-static-ip-name: "urlshortener-ip"
    networking.gke.io/managed-certificates: "urlshortener-cert"
    kubernetes.io/ingress.allow-http: "false"
spec:
  rules:
  - host: yourdomain.com
    http:
      paths:
      - path: /*
        pathType: ImplementationSpecific
        backend:
          service:
            name: urlshortener
            port:
              number: 80
---
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
metadata:
  name: urlshortener-cert
  namespace: production
spec:
  domains:
    - yourdomain.com
    - www.yourdomain.com
```

**5. Reserve Static IP:**

```bash
gcloud compute addresses create urlshortener-ip \
  --global \
  --ip-version IPV4

# Get IP address
gcloud compute addresses describe urlshortener-ip --global
```

**6. Apply Ingress:**

```bash
kubectl apply -f k8s/ingress.yaml
```

---

## Secrets Management

**1. Enable Secret Manager:**

```bash
gcloud services enable secretmanager.googleapis.com
```

**2. Create Secrets:**

```bash
# Database password
echo -n '<SECURE_DB_PASSWORD>' | \
  gcloud secrets create db-password --data-file=-

# Redis auth string
echo -n '<REDIS_AUTH_STRING>' | \
  gcloud secrets create redis-password --data-file=-

# JWT secret
echo -n '<JWT_SECRET>' | \
  gcloud secrets create jwt-secret --data-file=-
```

**3. Grant Access to GKE Service Account:**

```bash
# Get GKE service account
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
GSA_EMAIL="$PROJECT_NUMBER-compute@developer.gserviceaccount.com"

# Grant access
gcloud secrets add-iam-policy-binding db-password \
  --member="serviceAccount:$GSA_EMAIL" \
  --role="roles/secretmanager.secretAccessor"
```

---

## Monitoring and Logging

**1. View Logs:**

```bash
# Application logs
gcloud logging read "resource.type=k8s_container AND resource.labels.namespace_name=production" \
  --limit 50 \
  --format json

# Or use kubectl
kubectl logs -f deployment/urlshortener -n production
```

**2. Create Monitoring Dashboard:**

```bash
# Install Cloud Monitoring in GKE (already enabled)
# Access via Cloud Console > Monitoring > Dashboards
```

**3. Create Alerts:**

```bash
# High CPU alert
gcloud alpha monitoring policies create \
  --notification-channels=<CHANNEL_ID> \
  --display-name="High CPU Alert" \
  --condition-display-name="CPU > 80%" \
  --condition-threshold-value=0.8 \
  --condition-threshold-duration=300s \
  --condition-filter='resource.type="k8s_container" AND metric.type="kubernetes.io/container/cpu/core_usage_time"'
```

---

## Auto Scaling

**1. Horizontal Pod Autoscaler:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: urlshortener-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: urlshortener
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

```bash
kubectl apply -f k8s/hpa.yaml
```

**2. Cluster Autoscaler:**

Already enabled with `--enable-autoscaling` flag during cluster creation.

---

## Backup and Disaster Recovery

**1. Automated SQL Backups:**

```bash
# Backups are automatic with 7-day retention
# Create on-demand backup
gcloud sql backups create \
  --instance=urlshortener-db \
  --description="Manual backup before migration"
```

**2. Restore Database:**

```bash
# List backups
gcloud sql backups list --instance=urlshortener-db

# Restore
gcloud sql backups restore <BACKUP_ID> \
  --backup-instance=urlshortener-db \
  --backup-id=<BACKUP_ID>
```

**3. Export Database:**

```bash
# Create Cloud Storage bucket
gsutil mb gs://$PROJECT_ID-backups

# Export database
gcloud sql export sql urlshortener-db \
  gs://$PROJECT_ID-backups/backup-$(date +%Y%m%d).sql \
  --database=urlshortener_prod
```

---

## Cost Optimization

**Estimated Monthly Costs (us-central1):**

| Component | Size | Monthly Cost |
|-----------|------|--------------|
| GKE (e2-medium x2) | 2 nodes | $60 |
| Cloud SQL (custom-2-7680) | 7.5GB RAM | $100 |
| Memorystore (Standard 1GB) | HA | $80 |
| Cloud Load Balancing | Standard | $20 |
| Cloud CDN | 1TB transfer | $80 |
| Egress | 500GB | $65 |
| **Total** | | **~$405/month** |

**Cost Reduction:**
1. Use Committed Use Discounts (30-57% savings)
2. Use Preemptible VMs for non-critical workloads
3. Enable GKE cluster autoscaling
4. Use Cloud Storage Nearline for backups
5. Implement aggressive CDN caching
6. Use Cloud NAT instead of individual IPs

---

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n production
kubectl describe pod <POD_NAME> -n production
kubectl logs -f <POD_NAME> -n production
```

### Test Database Connection

```bash
# From inside pod
kubectl exec -it <POD_NAME> -n production -- /bin/sh
psql -h 127.0.0.1 -U appuser -d urlshortener_prod
```

### Test Redis Connection

```bash
kubectl exec -it <POD_NAME> -n production -- /bin/sh
redis-cli -h <REDIS_IP> -a <AUTH_STRING> ping
```

For more troubleshooting, see [TROUBLESHOOTING.md](../TROUBLESHOOTING.md).

---

**Last Updated:** 2025-12-19
**Maintained By:** DevOps Team
