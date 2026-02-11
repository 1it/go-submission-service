# Cloud Provider Deployment Guide

This guide covers deploying the Go Submission Service on major cloud platforms including AWS, Google Cloud Platform (GCP), and Microsoft Azure.

## Table of Contents

- [AWS Deployment](#aws-deployment)
- [Google Cloud Platform](#google-cloud-platform)
- [Microsoft Azure](#microsoft-azure)
- [Cloud-Native Features](#cloud-native-features)
- [Cost Optimization](#cost-optimization)
- [Multi-Cloud Considerations](#multi-cloud-considerations)

## AWS Deployment

### AWS ECS (Elastic Container Service)

#### Prerequisites

```bash
# Install AWS CLI
aws configure

# Install ECS CLI
ecs-cli configure --cluster form-submission --region us-west-2
```

#### Task Definition

```json
{
  "family": "form-submission-service",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "form-submission-service",
      "image": "ghcr.io/1it/go-submission-service:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENVIRONMENT",
          "value": "production"
        },
        {
          "name": "PORT",
          "value": "8080"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:ssm:us-west-2:ACCOUNT:parameter/form-service/database-url"
        },
        {
          "name": "MAILERSEND_API_KEY",
          "valueFrom": "arn:aws:ssm:us-west-2:ACCOUNT:parameter/form-service/mailersend-api-key"
        },
        {
          "name": "ADMIN_API_KEY",
          "valueFrom": "arn:aws:ssm:us-west-2:ACCOUNT:parameter/form-service/admin-api-key"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/form-submission-service",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost:8080/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

#### Service Definition

```json
{
  "serviceName": "form-submission-service",
  "cluster": "form-submission",
  "taskDefinition": "form-submission-service:1",
  "desiredCount": 2,
  "launchType": "FARGATE",
  "networkConfiguration": {
    "awsvpcConfiguration": {
      "subnets": [
        "subnet-12345678",
        "subnet-87654321"
      ],
      "securityGroups": [
        "sg-12345678"
      ],
      "assignPublicIp": "ENABLED"
    }
  },
  "loadBalancers": [
    {
      "targetGroupArn": "arn:aws:elasticloadbalancing:us-west-2:ACCOUNT:targetgroup/form-submission/1234567890123456",
      "containerName": "form-submission-service",
      "containerPort": 8080
    }
  ],
  "serviceRegistries": [
    {
      "registryArn": "arn:aws:servicediscovery:us-west-2:ACCOUNT:service/srv-12345678"
    }
  ]
}
```

#### Deployment Script

```bash
#!/bin/bash
# deploy-aws-ecs.sh

set -e

# Variables
CLUSTER_NAME="form-submission"
SERVICE_NAME="form-submission-service"
REGION="us-west-2"
IMAGE_TAG="latest"

# Create cluster
aws ecs create-cluster --cluster-name $CLUSTER_NAME --region $REGION

# Register task definition
aws ecs register-task-definition \
  --cli-input-json file://task-definition.json \
  --region $REGION

# Create service
aws ecs create-service \
  --cli-input-json file://service-definition.json \
  --region $REGION

# Wait for service to be stable
aws ecs wait services-stable \
  --cluster $CLUSTER_NAME \
  --services $SERVICE_NAME \
  --region $REGION

echo "Deployment completed successfully!"
```

### AWS RDS for Database

#### PostgreSQL Setup

```bash
# Create RDS instance
aws rds create-db-instance \
  --db-instance-identifier form-submission-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 14.9 \
  --master-username formservice \
  --master-user-password SecurePassword123! \
  --allocated-storage 20 \
  --storage-type gp2 \
  --vpc-security-group-ids sg-12345678 \
  --db-subnet-group-name default-vpc-12345678 \
  --backup-retention-period 7 \
  --storage-encrypted \
  --region us-west-2

# Get connection endpoint
aws rds describe-db-instances \
  --db-instance-identifier form-submission-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text
```

### AWS Application Load Balancer

```bash
# Create target group
aws elbv2 create-target-group \
  --name form-submission-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-12345678 \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3

# Create load balancer
aws elbv2 create-load-balancer \
  --name form-submission-alb \
  --subnets subnet-12345678 subnet-87654321 \
  --security-groups sg-12345678

# Create listener
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:us-west-2:ACCOUNT:loadbalancer/app/form-submission-alb/1234567890123456 \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=arn:aws:acm:us-west-2:ACCOUNT:certificate/12345678-1234-1234-1234-123456789012 \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:us-west-2:ACCOUNT:targetgroup/form-submission-tg/1234567890123456
```

### AWS CloudFormation Template

```yaml
# cloudformation-template.yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Form Submission Service Infrastructure'

Parameters:
  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC ID for the deployment
  
  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>
    Description: Subnet IDs for the deployment
  
  CertificateArn:
    Type: String
    Description: SSL Certificate ARN for HTTPS

Resources:
  # ECS Cluster
  ECSCluster:
    Type: AWS::ECS::Cluster
    Properties:
      ClusterName: form-submission
      CapacityProviders:
        - FARGATE
        - FARGATE_SPOT
      DefaultCapacityProviderStrategy:
        - CapacityProvider: FARGATE
          Weight: 1
        - CapacityProvider: FARGATE_SPOT
          Weight: 4

  # Security Group
  SecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for form submission service
      VpcId: !Ref VpcId
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 8080
          ToPort: 8080
          SourceSecurityGroupId: !Ref LoadBalancerSecurityGroup
      SecurityGroupEgress:
        - IpProtocol: -1
          CidrIp: 0.0.0.0/0

  # Load Balancer Security Group
  LoadBalancerSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for load balancer
      VpcId: !Ref VpcId
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0

  # Application Load Balancer
  LoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Name: form-submission-alb
      Scheme: internet-facing
      Type: application
      Subnets: !Ref SubnetIds
      SecurityGroups:
        - !Ref LoadBalancerSecurityGroup

  # Target Group
  TargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Name: form-submission-tg
      Port: 8080
      Protocol: HTTP
      VpcId: !Ref VpcId
      TargetType: ip
      HealthCheckPath: /health
      HealthCheckIntervalSeconds: 30
      HealthCheckTimeoutSeconds: 5
      HealthyThresholdCount: 2
      UnhealthyThresholdCount: 3

  # Listener
  Listener:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
        - Type: forward
          TargetGroupArn: !Ref TargetGroup
      LoadBalancerArn: !Ref LoadBalancer
      Port: 443
      Protocol: HTTPS
      Certificates:
        - CertificateArn: !Ref CertificateArn

  # ECS Task Definition
  TaskDefinition:
    Type: AWS::ECS::TaskDefinition
    Properties:
      Family: form-submission-service
      NetworkMode: awsvpc
      RequiresCompatibilities:
        - FARGATE
      Cpu: 512
      Memory: 1024
      ExecutionRoleArn: !Ref ExecutionRole
      TaskRoleArn: !Ref TaskRole
      ContainerDefinitions:
        - Name: form-submission-service
          Image: ghcr.io/1it/go-submission-service:latest
          PortMappings:
            - ContainerPort: 8080
          Environment:
            - Name: ENVIRONMENT
              Value: production
            - Name: PORT
              Value: "8080"
          Secrets:
            - Name: DATABASE_URL
              ValueFrom: !Ref DatabaseUrlParameter
            - Name: MAILERSEND_API_KEY
              ValueFrom: !Ref MailerSendApiKeyParameter
            - Name: ADMIN_API_KEY
              ValueFrom: !Ref AdminApiKeyParameter
          LogConfiguration:
            LogDriver: awslogs
            Options:
              awslogs-group: !Ref LogGroup
              awslogs-region: !Ref AWS::Region
              awslogs-stream-prefix: ecs
          HealthCheck:
            Command:
              - CMD-SHELL
              - curl -f http://localhost:8080/health || exit 1
            Interval: 30
            Timeout: 5
            Retries: 3
            StartPeriod: 60

  # ECS Service
  Service:
    Type: AWS::ECS::Service
    DependsOn: Listener
    Properties:
      ServiceName: form-submission-service
      Cluster: !Ref ECSCluster
      TaskDefinition: !Ref TaskDefinition
      DesiredCount: 2
      LaunchType: FARGATE
      NetworkConfiguration:
        AwsvpcConfiguration:
          Subnets: !Ref SubnetIds
          SecurityGroups:
            - !Ref SecurityGroup
          AssignPublicIp: ENABLED
      LoadBalancers:
        - TargetGroupArn: !Ref TargetGroup
          ContainerName: form-submission-service
          ContainerPort: 8080

  # IAM Roles
  ExecutionRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Statement:
          - Effect: Allow
            Principal:
              Service: ecs-tasks.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy
      Policies:
        - PolicyName: ParameterStoreAccess
          PolicyDocument:
            Statement:
              - Effect: Allow
                Action:
                  - ssm:GetParameters
                  - ssm:GetParameter
                Resource:
                  - !Ref DatabaseUrlParameter
                  - !Ref MailerSendApiKeyParameter
                  - !Ref AdminApiKeyParameter

  TaskRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Statement:
          - Effect: Allow
            Principal:
              Service: ecs-tasks.amazonaws.com
            Action: sts:AssumeRole

  # CloudWatch Log Group
  LogGroup:
    Type: AWS::Logs::LogGroup
    Properties:
      LogGroupName: /ecs/form-submission-service
      RetentionInDays: 30

  # SSM Parameters
  DatabaseUrlParameter:
    Type: AWS::SSM::Parameter
    Properties:
      Name: /form-service/database-url
      Type: SecureString
      Value: !Sub 'postgres://formservice:${DatabasePassword}@${DatabaseEndpoint}:5432/submissions'
      Description: Database connection URL

  MailerSendApiKeyParameter:
    Type: AWS::SSM::Parameter
    Properties:
      Name: /form-service/mailersend-api-key
      Type: SecureString
      Value: your-mailersend-api-key
      Description: MailerSend API key

  AdminApiKeyParameter:
    Type: AWS::SSM::Parameter
    Properties:
      Name: /form-service/admin-api-key
      Type: SecureString
      Value: !Sub '${AWS::StackName}-${AWS::AccountId}-admin-key'
      Description: Admin API key

Outputs:
  LoadBalancerDNS:
    Description: DNS name of the load balancer
    Value: !GetAtt LoadBalancer.DNSName
    Export:
      Name: !Sub '${AWS::StackName}-LoadBalancerDNS'

  ServiceArn:
    Description: ARN of the ECS service
    Value: !Ref Service
    Export:
      Name: !Sub '${AWS::StackName}-ServiceArn'
```

## Google Cloud Platform

### Google Cloud Run

#### Deployment

```bash
# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable secretmanager.googleapis.com

# Deploy to Cloud Run
gcloud run deploy form-submission-service \
  --image ghcr.io/1it/go-submission-service:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --memory 1Gi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --concurrency 80 \
  --timeout 300 \
  --set-env-vars ENVIRONMENT=production,PORT=8080 \
  --set-secrets DATABASE_URL=database-url:latest,MAILERSEND_API_KEY=mailersend-api-key:latest,ADMIN_API_KEY=admin-api-key:latest
```

#### Cloud Run YAML Configuration

```yaml
# service.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: form-submission-service
  annotations:
    run.googleapis.com/ingress: all
    run.googleapis.com/execution-environment: gen2
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "10"
        run.googleapis.com/cpu-throttling: "true"
        run.googleapis.com/execution-environment: gen2
    spec:
      containerConcurrency: 80
      timeoutSeconds: 300
      containers:
      - image: ghcr.io/1it/go-submission-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: production
        - name: PORT
          value: "8080"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-url
              key: latest
        - name: MAILERSEND_API_KEY
          valueFrom:
            secretKeyRef:
              name: mailersend-api-key
              key: latest
        - name: ADMIN_API_KEY
          valueFrom:
            secretKeyRef:
              name: admin-api-key
              key: latest
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        startupProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          failureThreshold: 3
```

#### Secret Manager Setup

```bash
# Create secrets
echo -n "postgres://user:pass@host:5432/db" | gcloud secrets create database-url --data-file=-
echo -n "your-mailersend-api-key" | gcloud secrets create mailersend-api-key --data-file=-
echo -n "your-admin-api-key" | gcloud secrets create admin-api-key --data-file=-

# Grant access to Cloud Run service account
gcloud secrets add-iam-policy-binding database-url \
  --member="serviceAccount:PROJECT-NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding mailersend-api-key \
  --member="serviceAccount:PROJECT-NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding admin-api-key \
  --member="serviceAccount:PROJECT-NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Google Cloud SQL

```bash
# Create Cloud SQL instance
gcloud sql instances create form-submission-db \
  --database-version=POSTGRES_14 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --storage-type=SSD \
  --storage-size=10GB \
  --backup-start-time=02:00 \
  --enable-bin-log \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=03 \
  --maintenance-release-channel=production

# Create database
gcloud sql databases create submissions --instance=form-submission-db

# Create user
gcloud sql users create formservice \
  --instance=form-submission-db \
  --password=SecurePassword123!

# Get connection name
gcloud sql instances describe form-submission-db --format="value(connectionName)"
```

### GKE Deployment

```yaml
# gke-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: form-submission-service
  labels:
    app: form-submission-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: form-submission-service
  template:
    metadata:
      labels:
        app: form-submission-service
    spec:
      containers:
      - name: form-submission-service
        image: ghcr.io/1it/go-submission-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: production
        - name: PORT
          value: "8080"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: database-url
        - name: MAILERSEND_API_KEY
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: mailersend-api-key
        - name: ADMIN_API_KEY
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: admin-api-key
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: form-submission-service
spec:
  selector:
    app: form-submission-service
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Microsoft Azure

### Azure Container Instances

```bash
# Create resource group
az group create --name form-submission-rg --location eastus

# Create container instance
az container create \
  --resource-group form-submission-rg \
  --name form-submission-service \
  --image ghcr.io/1it/go-submission-service:latest \
  --cpu 1 \
  --memory 1 \
  --ports 8080 \
  --dns-name-label form-submission-unique \
  --environment-variables ENVIRONMENT=production PORT=8080 \
  --secure-environment-variables DATABASE_URL="postgres://user:pass@host:5432/db" MAILERSEND_API_KEY="your-api-key" ADMIN_API_KEY="your-admin-key" \
  --restart-policy Always
```

### Azure Container Apps

```yaml
# container-app.yaml
apiVersion: 2022-03-01
type: Microsoft.App/containerApps
name: form-submission-service
location: East US
properties:
  managedEnvironmentId: /subscriptions/SUBSCRIPTION-ID/resourceGroups/form-submission-rg/providers/Microsoft.App/managedEnvironments/form-submission-env
  configuration:
    ingress:
      external: true
      targetPort: 8080
      allowInsecure: false
    secrets:
    - name: database-url
      value: postgres://user:pass@host:5432/db
    - name: mailersend-api-key
      value: your-mailersend-api-key
    - name: admin-api-key
      value: your-admin-api-key
  template:
    containers:
    - name: form-submission-service
      image: ghcr.io/1it/go-submission-service:latest
      env:
      - name: ENVIRONMENT
        value: production
      - name: PORT
        value: "8080"
      - name: DATABASE_URL
        secretRef: database-url
      - name: MAILERSEND_API_KEY
        secretRef: mailersend-api-key
      - name: ADMIN_API_KEY
        secretRef: admin-api-key
      resources:
        cpu: 0.5
        memory: 1Gi
      probes:
      - type: Liveness
        httpGet:
          path: /health
          port: 8080
        initialDelaySeconds: 30
        periodSeconds: 30
      - type: Readiness
        httpGet:
          path: /health
          port: 8080
        initialDelaySeconds: 5
        periodSeconds: 10
    scale:
      minReplicas: 1
      maxReplicas: 10
      rules:
      - name: http-scaling
        http:
          metadata:
            concurrentRequests: 50
```

### Azure Database for PostgreSQL

```bash
# Create PostgreSQL server
az postgres server create \
  --resource-group form-submission-rg \
  --name form-submission-db \
  --location eastus \
  --admin-user formservice \
  --admin-password SecurePassword123! \
  --sku-name B_Gen5_1 \
  --version 11

# Create database
az postgres db create \
  --resource-group form-submission-rg \
  --server-name form-submission-db \
  --name submissions

# Configure firewall
az postgres server firewall-rule create \
  --resource-group form-submission-rg \
  --server form-submission-db \
  --name AllowAzureServices \
  --start-ip-address 0.0.0.0 \
  --end-ip-address 0.0.0.0
```

### Azure Kubernetes Service (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group form-submission-rg \
  --name form-submission-aks \
  --node-count 3 \
  --node-vm-size Standard_B2s \
  --enable-addons monitoring \
  --generate-ssh-keys

# Get credentials
az aks get-credentials \
  --resource-group form-submission-rg \
  --name form-submission-aks

# Deploy application
kubectl apply -f aks-deployment.yaml
```

## Cloud-Native Features

### Auto-scaling Configuration

#### AWS ECS Auto Scaling

```json
{
  "ServiceName": "form-submission-service",
  "ClusterName": "form-submission",
  "ScalableDimension": "ecs:service:DesiredCount",
  "ServiceNamespace": "ecs",
  "PolicyName": "form-submission-scaling-policy",
  "PolicyType": "TargetTrackingScaling",
  "TargetTrackingScalingPolicyConfiguration": {
    "TargetValue": 70.0,
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
    },
    "ScaleOutCooldown": 300,
    "ScaleInCooldown": 300
  }
}
```

#### GCP Cloud Run Auto Scaling

```yaml
metadata:
  annotations:
    autoscaling.knative.dev/minScale: "1"
    autoscaling.knative.dev/maxScale: "100"
    autoscaling.knative.dev/target: "70"
```

#### Azure Container Apps Auto Scaling

```yaml
scale:
  minReplicas: 1
  maxReplicas: 20
  rules:
  - name: cpu-scaling
    custom:
      type: cpu
      metadata:
        type: Utilization
        value: "70"
  - name: memory-scaling
    custom:
      type: memory
      metadata:
        type: Utilization
        value: "80"
```

### Load Balancing

#### AWS Application Load Balancer

```bash
# Health check configuration
aws elbv2 modify-target-group \
  --target-group-arn arn:aws:elasticloadbalancing:region:account:targetgroup/form-submission-tg/id \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3
```

#### GCP Load Balancer

```bash
# Create global load balancer
gcloud compute url-maps create form-submission-lb \
  --default-service form-submission-backend

gcloud compute target-https-proxies create form-submission-proxy \
  --url-map form-submission-lb \
  --ssl-certificates form-submission-ssl

gcloud compute forwarding-rules create form-submission-forwarding-rule \
  --global \
  --target-https-proxy form-submission-proxy \
  --ports 443
```

#### Azure Application Gateway

```bash
# Create application gateway
az network application-gateway create \
  --name form-submission-gateway \
  --resource-group form-submission-rg \
  --location eastus \
  --capacity 2 \
  --sku Standard_v2 \
  --vnet-name form-submission-vnet \
  --subnet gateway-subnet \
  --public-ip-address form-submission-pip \
  --http-settings-cookie-based-affinity Disabled \
  --http-settings-port 8080 \
  --http-settings-protocol Http
```

### Monitoring and Logging

#### AWS CloudWatch

```json
{
  "logConfiguration": {
    "logDriver": "awslogs",
    "options": {
      "awslogs-group": "/ecs/form-submission-service",
      "awslogs-region": "us-west-2",
      "awslogs-stream-prefix": "ecs"
    }
  }
}
```

#### GCP Cloud Logging

```yaml
metadata:
  annotations:
    run.googleapis.com/logging: "json"
```

#### Azure Monitor

```bash
# Enable container insights
az aks enable-addons \
  --resource-group form-submission-rg \
  --name form-submission-aks \
  --addons monitoring
```

## Cost Optimization

### AWS Cost Optimization

1. **Use Spot Instances**:
   ```yaml
   CapacityProviders:
     - FARGATE
     - FARGATE_SPOT
   DefaultCapacityProviderStrategy:
     - CapacityProvider: FARGATE
       Weight: 1
     - CapacityProvider: FARGATE_SPOT
       Weight: 4
   ```

2. **Right-size Resources**:
   ```json
   {
     "cpu": "256",
     "memory": "512"
   }
   ```

3. **Use Reserved Capacity**:
   ```bash
   # Purchase Savings Plans for predictable workloads
   aws savingsplans create-savings-plan \
     --savings-plan-type Compute \
     --term-duration-in-years 1 \
     --payment-option All_Upfront \
     --commitment 10
   ```

### GCP Cost Optimization

1. **Use Preemptible Instances**:
   ```bash
   gcloud container node-pools create preemptible-pool \
     --cluster form-submission-aks \
     --preemptible \
     --num-nodes 3
   ```

2. **Cloud Run Concurrency**:
   ```yaml
   spec:
     containerConcurrency: 1000  # Maximize concurrency
   ```

3. **Committed Use Discounts**:
   ```bash
   # Purchase committed use contracts
   gcloud compute commitments create form-submission-commitment \
     --plan 12-month \
     --resources vcpus=4,memory=8GB
   ```

### Azure Cost Optimization

1. **Use Spot Instances**:
   ```bash
   az aks nodepool add \
     --cluster-name form-submission-aks \
     --name spotpool \
     --priority Spot \
     --eviction-policy Delete \
     --spot-max-price -1
   ```

2. **Reserved Instances**:
   ```bash
   # Purchase reserved instances
   az reservations reservation-order purchase \
     --reserved-resource-type VirtualMachines \
     --sku Standard_B2s \
     --location eastus \
     --term P1Y
   ```

## Multi-Cloud Considerations

### Infrastructure as Code

#### Terraform Multi-Cloud

```hcl
# main.tf
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

# AWS Provider
provider "aws" {
  region = var.aws_region
}

# GCP Provider
provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
}

# Azure Provider
provider "azurerm" {
  features {}
}

# Variables
variable "deployment_target" {
  description = "Target cloud provider"
  type        = string
  validation {
    condition     = contains(["aws", "gcp", "azure"], var.deployment_target)
    error_message = "Deployment target must be aws, gcp, or azure."
  }
}

# Conditional deployments
module "aws_deployment" {
  count  = var.deployment_target == "aws" ? 1 : 0
  source = "./modules/aws"
  
  # AWS-specific variables
}

module "gcp_deployment" {
  count  = var.deployment_target == "gcp" ? 1 : 0
  source = "./modules/gcp"
  
  # GCP-specific variables
}

module "azure_deployment" {
  count  = var.deployment_target == "azure" ? 1 : 0
  source = "./modules/azure"
  
  # Azure-specific variables
}
```

### CI/CD Pipeline

```yaml
# .github/workflows/multi-cloud-deploy.yml
name: Multi-Cloud Deployment

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      target_cloud:
        description: 'Target cloud provider'
        required: true
        default: 'aws'
        type: choice
        options:
        - aws
        - gcp
        - azure

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v2
    
    - name: Configure AWS credentials
      if: github.event.inputs.target_cloud == 'aws'
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2
    
    - name: Configure GCP credentials
      if: github.event.inputs.target_cloud == 'gcp'
      uses: google-github-actions/auth@v1
      with:
        credentials_json: ${{ secrets.GCP_SA_KEY }}
    
    - name: Configure Azure credentials
      if: github.event.inputs.target_cloud == 'azure'
      uses: azure/login@v1
      with:
        creds: ${{ secrets.AZURE_CREDENTIALS }}
    
    - name: Terraform Init
      run: terraform init
    
    - name: Terraform Plan
      run: |
        terraform plan \
          -var="deployment_target=${{ github.event.inputs.target_cloud }}" \
          -out=tfplan
    
    - name: Terraform Apply
      run: terraform apply tfplan
```

### Monitoring Across Clouds

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  # AWS targets
  - job_name: 'aws-form-service'
    static_configs:
      - targets: ['aws-lb-dns:443']
    scheme: https
    metrics_path: '/metrics'
  
  # GCP targets
  - job_name: 'gcp-form-service'
    static_configs:
      - targets: ['gcp-cloud-run-url']
    scheme: https
    metrics_path: '/metrics'
  
  # Azure targets
  - job_name: 'azure-form-service'
    static_configs:
      - targets: ['azure-container-app-url']
    scheme: https
    metrics_path: '/metrics'
```

## Next Steps

- [Production Setup Guide](production.md)
- [Docker Deployment](docker.md)
- [Kubernetes Deployment](kubernetes.md)
- [SSL/TLS Configuration](ssl-tls.md)