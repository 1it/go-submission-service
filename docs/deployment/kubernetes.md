# Kubernetes Deployment Guide

This guide covers deploying the Go Submission Service to Kubernetes clusters.

## Prerequisites

- Kubernetes cluster (v1.20+)
- `kubectl` configured to access your cluster
- Container registry access (Docker Hub, GitHub Container Registry, etc.)
- Persistent storage support (for database)

## Quick Start

### 1. Create Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: form-submission
  labels:
    name: form-submission
```

```bash
kubectl apply -f namespace.yaml
```

### 2. Create ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: form-submission-config
  namespace: form-submission
data:
  PORT: "8080"
  DB_PATH: "/data/submissions.db"
  EMAIL_SERVICE: "smtp"
  SMTP_HOST: "smtp.example.com"
  SMTP_PORT: "587"
  SMTP_FROM: "noreply@example.com"
  TEMPLATES_DIR: "/app/templates"
  CORS_ALLOWED_ORIGINS: "https://yourdomain.com"
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
```

### 3. Create Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: form-submission-secrets
  namespace: form-submission
type: Opaque
data:
  # Base64 encoded values
  SMTP_USERNAME: <base64-encoded-username>
  SMTP_PASSWORD: <base64-encoded-password>
  RECAPTCHA_SECRET_KEY: <base64-encoded-recaptcha-key>
  ADMIN_API_KEY: <base64-encoded-admin-key>
```

```bash
# Create secrets from command line
kubectl create secret generic form-submission-secrets \
  --from-literal=SMTP_USERNAME=your-username \
  --from-literal=SMTP_PASSWORD=your-password \
  --from-literal=RECAPTCHA_SECRET_KEY=your-recaptcha-key \
  --from-literal=ADMIN_API_KEY=your-admin-key \
  -n form-submission
```

### 4. Create Persistent Volume

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: form-submission-data
  namespace: form-submission
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard  # Adjust based on your cluster
```

### 5. Create Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: form-submission-service
  namespace: form-submission
  labels:
    app: form-submission-service
spec:
  replicas: 2
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
          name: http
        env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: form-submission-config
              key: PORT
        - name: DB_PATH
          valueFrom:
            configMapKeyRef:
              name: form-submission-config
              key: DB_PATH
        - name: EMAIL_SERVICE
          valueFrom:
            configMapKeyRef:
              name: form-submission-config
              key: EMAIL_SERVICE
        - name: SMTP_HOST
          valueFrom:
            configMapKeyRef:
              name: form-submission-config
              key: SMTP_HOST
        - name: SMTP_PORT
          valueFrom:
            configMapKeyRef:
              name: form-submission-config
              key: SMTP_PORT
        - name: SMTP_FROM
          valueFrom:
            configMapKeyRef:
              name: form-submission-config
              key: SMTP_FROM
        - name: SMTP_USERNAME
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: SMTP_USERNAME
        - name: SMTP_PASSWORD
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: SMTP_PASSWORD
        - name: RECAPTCHA_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: RECAPTCHA_SECRET_KEY
        - name: ADMIN_API_KEY
          valueFrom:
            secretKeyRef:
              name: form-submission-secrets
              key: ADMIN_API_KEY
        volumeMounts:
        - name: data
          mountPath: /data
        - name: templates
          mountPath: /app/templates
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
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: form-submission-data
      - name: templates
        configMap:
          name: form-submission-templates
          optional: true
```

### 6. Create Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: form-submission-service
  namespace: form-submission
  labels:
    app: form-submission-service
spec:
  selector:
    app: form-submission-service
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
```

### 7. Create Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: form-submission-ingress
  namespace: form-submission
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - forms.yourdomain.com
    secretName: form-submission-tls
  rules:
  - host: forms.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: form-submission-service
            port:
              number: 80
```

## Deploy All Resources

```bash
# Apply all configurations
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml
kubectl apply -f pvc.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# Check deployment status
kubectl get pods -n form-submission
kubectl get services -n form-submission
kubectl get ingress -n form-submission
```

## Production Configuration

### High Availability Setup

```yaml
# deployment-ha.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: form-submission-service
  namespace: form-submission
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - form-submission-service
              topologyKey: kubernetes.io/hostname
      # ... rest of the pod spec
```

### Horizontal Pod Autoscaler

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: form-submission-hpa
  namespace: form-submission
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: form-submission-service
  minReplicas: 2
  maxReplicas: 10
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

### Network Policies

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: form-submission-network-policy
  namespace: form-submission
spec:
  podSelector:
    matchLabels:
      app: form-submission-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to: []  # Allow all egress for SMTP and external APIs
    ports:
    - protocol: TCP
      port: 587  # SMTP
    - protocol: TCP
      port: 465  # SMTP SSL
    - protocol: TCP
      port: 443  # HTTPS
    - protocol: TCP
      port: 53   # DNS
    - protocol: UDP
      port: 53   # DNS
```

## Monitoring and Observability

### ServiceMonitor for Prometheus

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: form-submission-metrics
  namespace: form-submission
  labels:
    app: form-submission-service
spec:
  selector:
    matchLabels:
      app: form-submission-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Grafana Dashboard ConfigMap

```yaml
# grafana-dashboard.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: form-submission-dashboard
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
data:
  form-submission-dashboard.json: |
    {
      "dashboard": {
        "title": "Form Submission Service",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total{job=\"form-submission-service\"}[5m])"
              }
            ]
          }
        ]
      }
    }
```

## Database Options

### SQLite (Development)

For development or small deployments, use SQLite with persistent storage:

```yaml
volumeMounts:
- name: data
  mountPath: /data
volumes:
- name: data
  persistentVolumeClaim:
    claimName: form-submission-data
```

### PostgreSQL (Production)

For production, consider using PostgreSQL:

```yaml
# postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: form-submission
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_DB
          value: submissions
        - name: POSTGRES_USER
          value: formservice
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        ports:
        - containerPort: 5432
      volumes:
      - name: postgres-data
        persistentVolumeClaim:
          claimName: postgres-data
```

## Backup and Recovery

### Database Backup CronJob

```yaml
# backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: database-backup
  namespace: form-submission
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: alpine:latest
            command:
            - /bin/sh
            - -c
            - |
              apk add --no-cache sqlite
              DATE=$(date +%Y%m%d_%H%M%S)
              sqlite3 /data/submissions.db ".backup /backup/submissions_$DATE.db"
              # Upload to cloud storage here
            volumeMounts:
            - name: data
              mountPath: /data
              readOnly: true
            - name: backup
              mountPath: /backup
          volumes:
          - name: data
            persistentVolumeClaim:
              claimName: form-submission-data
          - name: backup
            persistentVolumeClaim:
              claimName: backup-storage
          restartPolicy: OnFailure
```

## Troubleshooting

### Common Issues

1. **Pod not starting**:
   ```bash
   kubectl describe pod <pod-name> -n form-submission
   kubectl logs <pod-name> -n form-submission
   ```

2. **Service not accessible**:
   ```bash
   kubectl get svc -n form-submission
   kubectl describe svc form-submission-service -n form-submission
   ```

3. **Ingress issues**:
   ```bash
   kubectl get ingress -n form-submission
   kubectl describe ingress form-submission-ingress -n form-submission
   ```

4. **Database connection issues**:
   ```bash
   kubectl exec -it <pod-name> -n form-submission -- ls -la /data
   kubectl exec -it <pod-name> -n form-submission -- sqlite3 /data/submissions.db ".tables"
   ```

### Debug Commands

```bash
# Check all resources
kubectl get all -n form-submission

# Check events
kubectl get events -n form-submission --sort-by='.lastTimestamp'

# Check pod logs
kubectl logs -f deployment/form-submission-service -n form-submission

# Execute commands in pod
kubectl exec -it deployment/form-submission-service -n form-submission -- /bin/sh

# Port forward for local testing
kubectl port-forward svc/form-submission-service 8080:80 -n form-submission
```

## Security Best Practices

1. **Use specific image tags**:
   ```yaml
   image: ghcr.io/1it/go-submission-service:v1.0.0
   ```

2. **Run as non-root user**:
   ```yaml
   securityContext:
     runAsNonRoot: true
     runAsUser: 1000
     fsGroup: 1000
   ```

3. **Use Pod Security Standards**:
   ```yaml
   metadata:
     labels:
       pod-security.kubernetes.io/enforce: restricted
   ```

4. **Limit resource usage**:
   ```yaml
   resources:
     limits:
       memory: "512Mi"
       cpu: "500m"
     requests:
       memory: "128Mi"
       cpu: "100m"
   ```

## Next Steps

- [Production Setup Guide](production.md)
- [SSL/TLS Configuration](ssl-tls.md)
- [Cloud Provider Guides](cloud-providers.md)
- [Monitoring Setup](../monitoring.md)