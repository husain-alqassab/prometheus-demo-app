# Prometheus Demo Application for OpenShift Developer Sandbox

A production-style demo application that exposes Prometheus metrics for monitoring in OpenShift Developer Sandbox.

## Features

- **Language**: Go 1.21 with Prometheus client library
- **Metrics Exposed**:
  - HTTP request count (with labels: method, endpoint, status)
  - Request latency histogram
  - Application uptime
  - Custom business metrics (processed_jobs_total, active_jobs)
- **Endpoints**:
  - `/` - Main application endpoint
  - `/metrics` - Prometheus metrics endpoint
  - `/health` - Health check endpoint
  - `/ready` - Readiness check endpoint
- **OpenShift Compatible**:
  - Runs as non-root user
  - No privileged operations
  - Lightweight distroless base image
  - Minimal resource usage

## Prerequisites

- OpenShift Developer Sandbox account
- `oc` CLI tool installed
- Docker installed (for local testing)

## Local Development

### Build and Run Locally

```bash
# Download dependencies
go mod download

# Run the application
go run main.go
```

Test the endpoints:
```bash
# Main endpoint
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Metrics
curl http://localhost:8080/metrics
```

### Build Docker Image Locally

```bash
# Build the image
docker build -t prometheus-demo-app:latest .

# Run the container
docker run -p 8080:8080 prometheus-demo-app:latest
```

## OpenShift Developer Sandbox Deployment

### 1. Login to OpenShift

```bash
oc login https://api.sandbox.x8i5.p1.openshiftapps.com:6443
```

### 2. Create a New Project

```bash
oc new-project prometheus-demo
```

### 3. Build the Image Using OpenShift BuildConfig

```bash
# Create a BuildConfig
oc new-build --binary --name=prometheus-demo-app -l app=prometheus-demo-app

# Start the build from local directory
oc start-build prometheus-demo-app --from-dir=. --follow
```

### 4. Deploy the Application

```bash
# Apply the deployment
oc apply -f deployment.yaml

# Apply the service
oc apply -f service.yaml

# Apply the route
oc apply -f route.yaml
```

### Alternative: Deploy Using DeploymentConfig (Simpler)

If you prefer using DeploymentConfig instead of Deployment:

```bash
# Create the application using oc new-app
oc new-app prometheus-demo-app:latest --name=prometheus-demo-app

# Expose the service
oc expose svc/prometheus-demo-app

# Add health checks
oc set probe deployment/prometheus-demo-app --readiness --get-url=http://:8080/ready
oc set probe deployment/prometheus-demo-app --liveness --get-url=http://:8080/health

# Add resource limits
oc set resources deployment/prometheus-demo-app --limits=cpu=200m,memory=128Mi --requests=cpu=100m,memory=64Mi
```

### 5. Verify Deployment

```bash
# Check pod status
oc get pods

# Check the route
oc get route

# Get the route URL
ROUTE_URL=$(oc get route prometheus-demo-app -o jsonpath='{.spec.host}')
echo "Route URL: https://$ROUTE_URL"

# Test the application
curl https://$ROUTE_URL/

# Test health endpoint
curl https://$ROUTE_URL/health

# Test metrics endpoint
curl https://$ROUTE_URL/metrics
```

### 6. Check Logs

```bash
# Get pod name
POD_NAME=$(oc get pods -l app=prometheus-demo-app -o jsonpath='{.items[0].metadata.name}')

# View logs
oc logs $POD_NAME

# Follow logs
oc logs -f $POD_NAME
```

## Prometheus Metrics Discovery

### How Prometheus Discovers and Scrapes the Application

1. **Service Annotations**: The Service has Prometheus annotations that enable automatic discovery:
   ```yaml
   annotations:
     prometheus.io/scrape: "true"
     prometheus.io/path: "/metrics"
     prometheus.io/port: "8080"
   ```

2. **Prometheus Operator**: In OpenShift, the Prometheus Operator uses these annotations to automatically discover and configure scrape targets.

3. **Scrape Configuration**: Prometheus will:
   - Discover the service via annotations
   - Scrape metrics from `http://<service>:8080/metrics`
   - Collect metrics at the configured interval (default 15s)

4. **Metrics Available**: The following metrics are exposed:
   - `http_requests_total` - Counter for HTTP requests
   - `http_request_duration_seconds` - Histogram for request latency
   - `app_start_time_seconds` - Gauge for application start time
   - `processed_jobs_total` - Counter for processed jobs
   - `active_jobs` - Gauge for active jobs

### Manual Prometheus Configuration

If you need to manually configure Prometheus, add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'prometheus-demo-app'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - prometheus-demo
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

## Example Prometheus Metrics Output

```prometheus
# HELP http_request_duration_seconds HTTP request latency in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.005"} 0
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.01"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.025"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.05"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.1"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.25"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="0.5"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="1"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="2.5"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="5"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="10"} 1
http_request_duration_seconds_bucket{endpoint="/",method="GET",le="+Inf"} 1
http_request_duration_seconds_sum{endpoint="/",method="GET"} 0.012345678
http_request_duration_seconds_count{endpoint="/",method="GET"} 1

# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/",method="GET",status="200"} 5
http_requests_total{endpoint="/health",method="GET",status="200"} 2
http_requests_total{endpoint="/ready",method="GET",status="200"} 2

# HELP active_jobs Number of currently active jobs
# TYPE active_jobs gauge
active_jobs 0

# HELP app_start_time_seconds Application start time in Unix timestamp
# TYPE app_start_time_seconds gauge
app_start_time_seconds 1716472800

# HELP processed_jobs_total Total number of jobs processed
# TYPE processed_jobs_total counter
processed_jobs_total 5
```

## Troubleshooting

### Image Pull Errors

If you encounter image pull errors, ensure you've built the image in OpenShift:

```bash
# Check available image streams
oc get is

# Rebuild if needed
oc start-build prometheus-demo-app --from-dir=. --follow
```

### Pod Not Starting

Check pod status and logs:
```bash
oc get pods
oc describe pod <pod-name>
oc logs <pod-name>
```

### Metrics Not Being Scraped

1. Verify Service annotations:
```bash
oc get svc prometheus-demo-app -o yaml
```

2. Check if Prometheus is running in your cluster:
```bash
oc get pods -n openshift-monitoring
```

3. Verify metrics endpoint is accessible:
```bash
oc port-forward svc/prometheus-demo-app 8080:8080
curl http://localhost:8080/metrics
```

## Resource Optimization

The application is configured for minimal resource usage suitable for Developer Sandbox:
- Memory Request: 64Mi
- Memory Limit: 128Mi
- CPU Request: 100m
- CPU Limit: 200m

## Security Features

- Runs as non-root user (UID 65534)
- No privileged operations
- Read-only root filesystem
- All capabilities dropped
- Lightweight distroless base image

## Cleanup

```bash
# Delete all resources
oc delete route prometheus-demo-app
oc delete service prometheus-demo-app
oc delete deployment prometheus-demo-app
oc delete project prometheus-demo
```

## Bonus Features

- **Grafana Dashboard**: See `grafana-dashboard.json` for an example dashboard
- **ServiceMonitor**: See `servicemonitor.yaml` for OpenShift monitoring stack integration
- **RED Metrics**: The application implements Rate (request count), Errors (via status labels), and Duration (latency histogram) metrics