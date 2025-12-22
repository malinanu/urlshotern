# AWS Deployment Guide

This guide covers deploying the URL Shortener application to Amazon Web Services (AWS) with production-ready infrastructure.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Infrastructure Setup](#infrastructure-setup)
  - [VPC and Networking](#vpc-and-networking)
  - [RDS PostgreSQL](#rds-postgresql)
  - [ElastiCache Redis](#elasticache-redis)
  - [EC2 Instances](#ec2-instances)
  - [Application Load Balancer](#application-load-balancer)
  - [CloudFront CDN](#cloudfront-cdn)
- [Application Deployment](#application-deployment)
- [Secrets Management](#secrets-management)
- [Monitoring and Logging](#monitoring-and-logging)
- [Auto Scaling](#auto-scaling)
- [Backup and Disaster Recovery](#backup-and-disaster-recovery)
- [Cost Optimization](#cost-optimization)
- [Troubleshooting](#troubleshooting)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         CloudFront CDN                       │
│                    (Global Edge Locations)                   │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                 Application Load Balancer                    │
│                     (Multi-AZ, SSL/TLS)                      │
└─────────┬──────────────────────────────────┬────────────────┘
          │                                  │
┌─────────▼─────────┐              ┌────────▼──────────┐
│   EC2 Auto Scaling│              │  EC2 Auto Scaling │
│   Group (AZ-1)    │              │   Group (AZ-2)    │
│   ┌──────────┐    │              │   ┌──────────┐    │
│   │ App      │    │              │   │ App      │    │
│   │ Instance │    │              │   │ Instance │    │
│   └──────────┘    │              │   └──────────┘    │
└──────┬────────────┘              └────────┬──────────┘
       │                                    │
       │      ┌─────────────────────┐       │
       └──────▶   ElastiCache Redis ◀───────┘
              │    (Cluster Mode)    │
              └─────────────────────┘
                       │
              ┌────────▼──────────┐
              │   RDS PostgreSQL  │
              │   (Multi-AZ)      │
              └───────────────────┘
```

**Key Components:**
- **CloudFront**: Global CDN for static assets and URL redirects
- **ALB**: Application Load Balancer with SSL termination
- **EC2 Auto Scaling**: Horizontally scalable compute instances
- **RDS PostgreSQL**: Managed database with automatic backups
- **ElastiCache Redis**: Managed Redis cluster for caching
- **Secrets Manager**: Secure storage for credentials
- **CloudWatch**: Logging, metrics, and monitoring
- **Route 53**: DNS management
- **S3**: Static asset storage and backups

---

## Prerequisites

### Required Tools
- AWS CLI v2: `aws --version`
- Terraform v1.0+ (optional, for IaC): `terraform --version`
- Docker: `docker --version`
- PostgreSQL client: `psql --version`

### AWS Account Requirements
- AWS account with administrative access
- AWS CLI configured with credentials: `aws configure`
- Budget allocation (estimated $100-300/month for small-medium deployment)

### Domain Setup
- Registered domain name (can use Route 53 or external registrar)
- Access to DNS management
- SSL certificate (can use ACM for free)

---

## Infrastructure Setup

### VPC and Networking

**1. Create VPC**

```bash
# Create VPC
aws ec2 create-vpc \
  --cidr-block 10.0.0.0/16 \
  --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=urlshortener-vpc}]'

# Enable DNS hostnames
aws ec2 modify-vpc-attribute \
  --vpc-id vpc-xxxxx \
  --enable-dns-hostnames
```

**2. Create Subnets (Multi-AZ)**

```bash
# Public subnet AZ-1
aws ec2 create-subnet \
  --vpc-id vpc-xxxxx \
  --cidr-block 10.0.1.0/24 \
  --availability-zone us-east-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=public-subnet-1a}]'

# Public subnet AZ-2
aws ec2 create-subnet \
  --vpc-id vpc-xxxxx \
  --cidr-block 10.0.2.0/24 \
  --availability-zone us-east-1b \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=public-subnet-1b}]'

# Private subnet AZ-1 (for RDS, ElastiCache)
aws ec2 create-subnet \
  --vpc-id vpc-xxxxx \
  --cidr-block 10.0.10.0/24 \
  --availability-zone us-east-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=private-subnet-1a}]'

# Private subnet AZ-2
aws ec2 create-subnet \
  --vpc-id vpc-xxxxx \
  --cidr-block 10.0.11.0/24 \
  --availability-zone us-east-1b \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=private-subnet-1b}]'
```

**3. Internet Gateway**

```bash
# Create and attach Internet Gateway
aws ec2 create-internet-gateway \
  --tag-specifications 'ResourceType=internet-gateway,Tags=[{Key=Name,Value=urlshortener-igw}]'

aws ec2 attach-internet-gateway \
  --vpc-id vpc-xxxxx \
  --internet-gateway-id igw-xxxxx
```

**4. Route Tables**

```bash
# Public route table
aws ec2 create-route-table \
  --vpc-id vpc-xxxxx \
  --tag-specifications 'ResourceType=route-table,Tags=[{Key=Name,Value=public-rt}]'

# Add route to Internet Gateway
aws ec2 create-route \
  --route-table-id rtb-xxxxx \
  --destination-cidr-block 0.0.0.0/0 \
  --gateway-id igw-xxxxx

# Associate with public subnets
aws ec2 associate-route-table \
  --subnet-id subnet-xxxxx \
  --route-table-id rtb-xxxxx
```

**5. Security Groups**

```bash
# ALB Security Group (allow HTTP/HTTPS from internet)
aws ec2 create-security-group \
  --group-name alb-sg \
  --description "Security group for ALB" \
  --vpc-id vpc-xxxxx

aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxx \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0

# EC2 Security Group (allow traffic from ALB)
aws ec2 create-security-group \
  --group-name ec2-app-sg \
  --description "Security group for EC2 instances" \
  --vpc-id vpc-xxxxx

aws ec2 authorize-security-group-ingress \
  --group-id sg-yyyyy \
  --protocol tcp \
  --port 8080 \
  --source-group sg-xxxxx

# RDS Security Group (allow PostgreSQL from EC2)
aws ec2 create-security-group \
  --group-name rds-sg \
  --description "Security group for RDS" \
  --vpc-id vpc-xxxxx

aws ec2 authorize-security-group-ingress \
  --group-id sg-zzzzz \
  --protocol tcp \
  --port 5432 \
  --source-group sg-yyyyy

# ElastiCache Security Group (allow Redis from EC2)
aws ec2 create-security-group \
  --group-name elasticache-sg \
  --description "Security group for ElastiCache" \
  --vpc-id vpc-xxxxx

aws ec2 authorize-security-group-ingress \
  --group-id sg-wwwww \
  --protocol tcp \
  --port 6379 \
  --source-group sg-yyyyy
```

---

### RDS PostgreSQL

**1. Create DB Subnet Group**

```bash
aws rds create-db-subnet-group \
  --db-subnet-group-name urlshortener-db-subnet \
  --db-subnet-group-description "Subnet group for URL Shortener DB" \
  --subnet-ids subnet-private-1a subnet-private-1b \
  --tags Key=Name,Value=urlshortener-db-subnet
```

**2. Create RDS Instance**

```bash
aws rds create-db-instance \
  --db-instance-identifier urlshortener-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 14.7 \
  --master-username dbadmin \
  --master-user-password <SECURE_PASSWORD> \
  --allocated-storage 20 \
  --storage-type gp3 \
  --storage-encrypted \
  --db-subnet-group-name urlshortener-db-subnet \
  --vpc-security-group-ids sg-zzzzz \
  --backup-retention-period 7 \
  --preferred-backup-window "03:00-04:00" \
  --preferred-maintenance-window "mon:04:00-mon:05:00" \
  --multi-az \
  --publicly-accessible false \
  --enable-performance-insights \
  --tags Key=Name,Value=urlshortener-db
```

**Instance Sizing Recommendations:**
| Workload | Instance Class | Storage | Cost/Month |
|----------|---------------|---------|------------|
| Dev/Test | db.t3.micro | 20GB | ~$20 |
| Small Production | db.t3.small | 50GB | ~$50 |
| Medium Production | db.t3.medium | 100GB | ~$100 |
| Large Production | db.m5.large | 500GB | ~$300 |

**3. Initialize Database**

```bash
# Get RDS endpoint
aws rds describe-db-instances \
  --db-instance-identifier urlshortener-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text

# Connect from bastion host or local (if whitelisted)
psql -h <RDS_ENDPOINT> -U dbadmin -d postgres

# Create database
CREATE DATABASE urlshortener_prod;
\c urlshortener_prod

# Run schema migrations
psql -h <RDS_ENDPOINT> -U dbadmin -d urlshortener_prod -f configs/database.sql
```

---

### ElastiCache Redis

**1. Create Cache Subnet Group**

```bash
aws elasticache create-cache-subnet-group \
  --cache-subnet-group-name urlshortener-redis-subnet \
  --cache-subnet-group-description "Subnet group for Redis" \
  --subnet-ids subnet-private-1a subnet-private-1b
```

**2. Create Redis Cluster**

```bash
aws elasticache create-replication-group \
  --replication-group-id urlshortener-redis \
  --replication-group-description "Redis cluster for URL Shortener" \
  --engine redis \
  --engine-version 7.0 \
  --cache-node-type cache.t3.micro \
  --num-cache-clusters 2 \
  --automatic-failover-enabled \
  --multi-az-enabled \
  --cache-subnet-group-name urlshortener-redis-subnet \
  --security-group-ids sg-wwwww \
  --at-rest-encryption-enabled \
  --transit-encryption-enabled \
  --auth-token <SECURE_REDIS_PASSWORD> \
  --tags Key=Name,Value=urlshortener-redis
```

**Instance Sizing:**
| Workload | Node Type | Nodes | Cost/Month |
|----------|-----------|-------|------------|
| Dev/Test | cache.t3.micro | 1 | ~$15 |
| Small Production | cache.t3.small | 2 | ~$50 |
| Medium Production | cache.t3.medium | 2 | ~$100 |
| Large Production | cache.m5.large | 3 | ~$300 |

---

### EC2 Instances

**1. Create IAM Role for EC2**

```bash
# Create IAM role
aws iam create-role \
  --role-name urlshortener-ec2-role \
  --assume-role-policy-document file://ec2-trust-policy.json

# ec2-trust-policy.json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "ec2.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}

# Attach policies
aws iam attach-role-policy \
  --role-name urlshortener-ec2-role \
  --policy-arn arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy

aws iam attach-role-policy \
  --role-name urlshortener-ec2-role \
  --policy-arn arn:aws:iam::aws:policy/SecretsManagerReadWrite

# Create instance profile
aws iam create-instance-profile \
  --instance-profile-name urlshortener-ec2-profile

aws iam add-role-to-instance-profile \
  --instance-profile-name urlshortener-ec2-profile \
  --role-name urlshortener-ec2-role
```

**2. Create Launch Template**

```bash
# Build Docker image and push to ECR
aws ecr create-repository --repository-name urlshortener

aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com

docker build -t urlshortener .
docker tag urlshortener:latest <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/urlshortener:latest
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/urlshortener:latest
```

**User Data Script (user-data.sh):**
```bash
#!/bin/bash
set -e

# Install Docker
yum update -y
yum install -y docker
systemctl start docker
systemctl enable docker

# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com

# Fetch secrets from Secrets Manager
export DB_PASSWORD=$(aws secretsmanager get-secret-value \
  --secret-id prod/urlshortener/db-password \
  --query SecretString --output text)

export REDIS_PASSWORD=$(aws secretsmanager get-secret-value \
  --secret-id prod/urlshortener/redis-password \
  --query SecretString --output text)

export JWT_SECRET=$(aws secretsmanager get-secret-value \
  --secret-id prod/urlshortener/jwt-secret \
  --query SecretString --output text)

# Run application container
docker run -d \
  --name urlshortener \
  --restart unless-stopped \
  -p 8080:8080 \
  -e DB_HOST=<RDS_ENDPOINT> \
  -e DB_PASSWORD="$DB_PASSWORD" \
  -e REDIS_HOST=<ELASTICACHE_ENDPOINT> \
  -e REDIS_PASSWORD="$REDIS_PASSWORD" \
  -e JWT_SECRET="$JWT_SECRET" \
  -e ENVIRONMENT=production \
  <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/urlshortener:latest

# Start CloudWatch agent
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config \
  -m ec2 \
  -s \
  -c ssm:cloudwatch-config
```

**Create Launch Template:**
```bash
aws ec2 create-launch-template \
  --launch-template-name urlshortener-lt \
  --version-description "Initial version" \
  --launch-template-data '{
    "ImageId": "ami-0c55b159cbfafe1f0",
    "InstanceType": "t3.small",
    "IamInstanceProfile": {"Name": "urlshortener-ec2-profile"},
    "SecurityGroupIds": ["sg-yyyyy"],
    "UserData": "<BASE64_ENCODED_USER_DATA>",
    "TagSpecifications": [{
      "ResourceType": "instance",
      "Tags": [{"Key": "Name", "Value": "urlshortener-app"}]
    }]
  }'
```

---

### Application Load Balancer

**1. Create Target Group**

```bash
aws elbv2 create-target-group \
  --name urlshortener-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-xxxxx \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3 \
  --matcher HttpCode=200
```

**2. Request SSL Certificate (ACM)**

```bash
# Request certificate
aws acm request-certificate \
  --domain-name yourdomain.com \
  --subject-alternative-names www.yourdomain.com \
  --validation-method DNS

# Get validation records
aws acm describe-certificate \
  --certificate-arn arn:aws:acm:us-east-1:xxxxx:certificate/xxxxx

# Add CNAME records to Route 53 or your DNS provider
# Wait for validation (can take 5-30 minutes)
```

**3. Create Application Load Balancer**

```bash
aws elbv2 create-load-balancer \
  --name urlshortener-alb \
  --subnets subnet-public-1a subnet-public-1b \
  --security-groups sg-xxxxx \
  --scheme internet-facing \
  --type application \
  --ip-address-type ipv4 \
  --tags Key=Name,Value=urlshortener-alb
```

**4. Create Listeners**

```bash
# HTTPS listener (primary)
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:us-east-1:xxxxx:loadbalancer/app/urlshortener-alb/xxxxx \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=arn:aws:acm:us-east-1:xxxxx:certificate/xxxxx \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:us-east-1:xxxxx:targetgroup/urlshortener-tg/xxxxx

# HTTP listener (redirect to HTTPS)
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:us-east-1:xxxxx:loadbalancer/app/urlshortener-alb/xxxxx \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=redirect,RedirectConfig='{Protocol=HTTPS,Port=443,StatusCode=HTTP_301}'
```

---

### CloudFront CDN

**1. Create S3 Bucket for Static Assets**

```bash
aws s3 mb s3://urlshortener-static-assets
aws s3 sync frontend/out s3://urlshortener-static-assets
```

**2. Create CloudFront Distribution**

```bash
# Create distribution configuration
aws cloudfront create-distribution \
  --distribution-config file://cloudfront-config.json
```

**cloudfront-config.json:**
```json
{
  "CallerReference": "urlshortener-2025-12-19",
  "Comment": "CDN for URL Shortener",
  "Enabled": true,
  "Origins": [
    {
      "Id": "alb-origin",
      "DomainName": "urlshortener-alb-xxxxx.us-east-1.elb.amazonaws.com",
      "CustomOriginConfig": {
        "HTTPPort": 80,
        "HTTPSPort": 443,
        "OriginProtocolPolicy": "https-only",
        "OriginSslProtocols": {"Items": ["TLSv1.2"], "Quantity": 1}
      }
    },
    {
      "Id": "s3-origin",
      "DomainName": "urlshortener-static-assets.s3.amazonaws.com",
      "S3OriginConfig": {
        "OriginAccessIdentity": ""
      }
    }
  ],
  "DefaultCacheBehavior": {
    "TargetOriginId": "alb-origin",
    "ViewerProtocolPolicy": "redirect-to-https",
    "AllowedMethods": {"Items": ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"], "Quantity": 7},
    "CachedMethods": {"Items": ["GET", "HEAD"], "Quantity": 2},
    "ForwardedValues": {
      "QueryString": true,
      "Cookies": {"Forward": "all"},
      "Headers": {"Items": ["Authorization", "Host"], "Quantity": 2}
    },
    "MinTTL": 0,
    "DefaultTTL": 86400,
    "MaxTTL": 31536000,
    "Compress": true
  },
  "CacheBehaviors": {
    "Quantity": 1,
    "Items": [{
      "PathPattern": "/_next/*",
      "TargetOriginId": "s3-origin",
      "ViewerProtocolPolicy": "redirect-to-https",
      "AllowedMethods": {"Items": ["GET", "HEAD"], "Quantity": 2},
      "MinTTL": 31536000,
      "DefaultTTL": 31536000,
      "MaxTTL": 31536000,
      "Compress": true
    }]
  },
  "Aliases": {"Items": ["yourdomain.com", "www.yourdomain.com"], "Quantity": 2},
  "ViewerCertificate": {
    "ACMCertificateArn": "arn:aws:acm:us-east-1:xxxxx:certificate/xxxxx",
    "SSLSupportMethod": "sni-only",
    "MinimumProtocolVersion": "TLSv1.2_2021"
  }
}
```

---

## Application Deployment

### Auto Scaling Group

```bash
aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name urlshortener-asg \
  --launch-template LaunchTemplateName=urlshortener-lt,Version='$Latest' \
  --min-size 2 \
  --max-size 10 \
  --desired-capacity 2 \
  --default-cooldown 300 \
  --health-check-type ELB \
  --health-check-grace-period 300 \
  --vpc-zone-identifier "subnet-public-1a,subnet-public-1b" \
  --target-group-arns arn:aws:elasticloadbalancing:us-east-1:xxxxx:targetgroup/urlshortener-tg/xxxxx \
  --tags Key=Name,Value=urlshortener-app,PropagateAtLaunch=true
```

### Scaling Policies

```bash
# Scale up when CPU > 70%
aws autoscaling put-scaling-policy \
  --auto-scaling-group-name urlshortener-asg \
  --policy-name scale-up \
  --policy-type TargetTrackingScaling \
  --target-tracking-configuration '{
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ASGAverageCPUUtilization"
    },
    "TargetValue": 70.0
  }'

# Scale based on request count
aws autoscaling put-scaling-policy \
  --auto-scaling-group-name urlshortener-asg \
  --policy-name scale-on-requests \
  --policy-type TargetTrackingScaling \
  --target-tracking-configuration '{
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ALBRequestCountPerTarget",
      "ResourceLabel": "app/urlshortener-alb/xxxxx/targetgroup/urlshortener-tg/xxxxx"
    },
    "TargetValue": 1000.0
  }'
```

---

## Secrets Management

Store sensitive credentials in AWS Secrets Manager:

```bash
# Database password
aws secretsmanager create-secret \
  --name prod/urlshortener/db-password \
  --secret-string "<SECURE_DB_PASSWORD>"

# Redis password
aws secretsmanager create-secret \
  --name prod/urlshortener/redis-password \
  --secret-string "<SECURE_REDIS_PASSWORD>"

# JWT secret
aws secretsmanager create-secret \
  --name prod/urlshortener/jwt-secret \
  --secret-string "<SECURE_JWT_SECRET>"

# Google OAuth credentials
aws secretsmanager create-secret \
  --name prod/urlshortener/google-oauth \
  --secret-string '{"client_id":"xxx","client_secret":"yyy"}'
```

**Rotation Configuration:**
```bash
aws secretsmanager rotate-secret \
  --secret-id prod/urlshortener/db-password \
  --rotation-lambda-arn arn:aws:lambda:us-east-1:xxxxx:function:rotate-secret \
  --rotation-rules AutomaticallyAfterDays=90
```

---

## Monitoring and Logging

### CloudWatch Dashboards

```bash
# Create custom dashboard
aws cloudwatch put-dashboard \
  --dashboard-name urlshortener \
  --dashboard-body file://dashboard.json
```

### CloudWatch Alarms

```bash
# High CPU alarm
aws cloudwatch put-metric-alarm \
  --alarm-name urlshortener-high-cpu \
  --alarm-description "Alert when CPU exceeds 80%" \
  --metric-name CPUUtilization \
  --namespace AWS/EC2 \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:us-east-1:xxxxx:alerts

# Database connections alarm
aws cloudwatch put-metric-alarm \
  --alarm-name rds-high-connections \
  --alarm-description "Alert when DB connections exceed 80% of max" \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --dimensions Name=DBInstanceIdentifier,Value=urlshortener-db \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:us-east-1:xxxxx:alerts
```

---

## Backup and Disaster Recovery

### Automated RDS Backups

RDS automatic backups are configured with 7-day retention. To restore:

```bash
# List available backups
aws rds describe-db-snapshots \
  --db-instance-identifier urlshortener-db

# Restore from snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier urlshortener-db-restored \
  --db-snapshot-identifier rds:urlshortener-db-2025-12-19-03-00
```

### Manual Backups

```bash
# Create manual snapshot
aws rds create-db-snapshot \
  --db-instance-identifier urlshortener-db \
  --db-snapshot-identifier manual-backup-2025-12-19
```

---

## Cost Optimization

**Estimated Monthly Costs (us-east-1):**

| Component | Size | Monthly Cost |
|-----------|------|--------------|
| EC2 (t3.small x2) | 2 instances | $30 |
| RDS (db.t3.small Multi-AZ) | 50GB | $70 |
| ElastiCache (cache.t3.small x2) | 2 nodes | $50 |
| ALB | Standard | $20 |
| CloudFront | 1TB transfer | $85 |
| Data Transfer | 500GB | $45 |
| **Total** | | **~$300/month** |

**Cost Reduction Strategies:**
1. Use Reserved Instances (40-60% savings)
2. Enable RDS storage auto-scaling
3. Use S3 Intelligent-Tiering for backups
4. Implement CloudFront caching aggressively
5. Use Spot Instances for non-critical workloads
6. Enable Cost Explorer and set budgets

---

## Troubleshooting

### Check Application Logs

```bash
# SSH to EC2 instance
aws ssm start-session --target i-xxxxx

# View application logs
docker logs urlshortener

# CloudWatch Logs
aws logs tail /aws/ec2/urlshortener --follow
```

### Database Connection Issues

```bash
# Test from EC2 instance
psql -h <RDS_ENDPOINT> -U dbadmin -d urlshortener_prod -c "SELECT 1"

# Check security groups
aws ec2 describe-security-groups --group-ids sg-zzzzz
```

### Redis Connection Issues

```bash
# Test from EC2 instance
redis-cli -h <ELASTICACHE_ENDPOINT> -a <PASSWORD> --tls ping
```

For more troubleshooting, see [TROUBLESHOOTING.md](../TROUBLESHOOTING.md).

---

**Last Updated:** 2025-12-19
**Maintained By:** DevOps Team
