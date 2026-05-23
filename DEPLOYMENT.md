# OpenShift Developer Sandbox Deployment Guide

This guide provides step-by-step instructions for deploying the Prometheus Demo Application to OpenShift Developer Sandbox.

## Prerequisites

1. OpenShift Developer Sandbox account (free at https://developers.redhat.com/developer-sandbox)
2. `oc` CLI tool installed
3. Docker installed (for local testing)

## Step 1: Install and Configure oc CLI

### Download oc CLI

Visit: https://console.openshift.com/download-tool

Select your platform and download the `oc` binary.

### Login to OpenShift

```bash
oc login https://api.sandbox.x8i5.p1.openshiftapps.com:6443
```

You'll be prompted for your username and password (use your Red Hat Developer account credentials).

## Step 2: Create a New Project

```bash
oc new-project prometheus-demo
```

Or use an existing project:
```bash
oc project <your-project-name>
```

## Step 3: Build and Deploy the Application

### Option A: Using OpenShift BuildConfig (Recommended)

1. **Create a BuildConfig:**
```bash
oc new-build --binary --name=prometheus-demo-app -l app=prometheus-demo-app
```

2. **Start the build from local directory:**
```bash
oc start-build prometheus-demo-app --from-dir=. --follow
```

This will:
- Upload your source code to OpenShift
- Build the Docker image using S2I (Source-to-Image)
- Store the image in OpenShift's internal registry

3. **Deploy the application:**
```bash
oc apply -f deployment.yaml
oc apply -f service.yaml
oc apply -f route.yaml
```

### Option B: Using oc new-app (Simpler)

```bash
# Create deployment from the built image
oc new-app prometheus-demo-app:latest --name=prometheus-demo-app

# Expose the service
oc expose svc/prometheus-demo-app

# Add health checks
oc set probe deployment/prometheus-demo-app --readiness --get-url=http://:8080/ready
oc set probe deployment/prometheus-demo-app --liveness --get-url=http://:8080/health

# Add resource limits
oc set resources deployment/prometheus-demo-app --limits=cpu=200m,memory=128Mi --requests=cpu=100m,memory=64Mi
```

### Option C: Using Makefile

```bash
make deploy
```

## Step 4: Verify Deployment

### Check Pod Status

```bash
oc get pods
```

Expected output:
```
NAME                                  READY   STATUS    RESTARTS   AGE
prometheus-demo-app-xxxxxxxxx-xxxxx   1/1     Running   0          1m
```

### Check Service and Route

```bash
oc get svc,route
```

Expected output:
```
NAME                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/prometheus-demo-app  ClusterIP   172.30.xxx.xxx   <none>        8080/TCP   1m

NAME                                        HOST/PORT                                                      ...
route.route.openshift.io/prometheus-demo-app prometheus-demo-app-prometheus-demo.apps.sandbox.x8i5.p1...   ...
```

### Get the Application URL

```bash
oc get route prometheus-demo-app -o jsonpath='{.spec.host}'
```

### Test the Application

```bash
# Get the route URL
ROUTE_URL=$(oc get route prometheus-demo-app -o jsonpath='{.spec.host}')

# Test main endpoint
curl https://$ROUTE_URL/

# Test health endpoint
curl https://$ROUTE_URL/health

# Test readiness endpoint
curl https://$ROUTE_URL/ready

# Test metrics endpoint
curl https://$ROUTE_URL/metrics
```

## Step 5: Generate Some Traffic

To see meaningful metrics, generate some traffic:

```bash
ROUTE_URL=$(oc get route prometheus-demo-app -o jsonpath='{.spec.host}')

# Generate some requests
for i in {1..10}; do curl https://$ROUTE_URL/; done
for i in {1..5}; do curl https://$ROUTE_URL/health; done
for i in {1..5}; do curl https://$ROUTE_URL/ready; done
```

## Step 6: Verify Metrics

```bash
ROUTE_URL=$(oc get route prometheus-demo-app -o jsonpath='{.spec.host}')
curl https://$ROUTE_URL/metrics
```

You should see Prometheus metrics like:
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/",method="GET",status="200"} 10
http_requests_total{endpoint="/health",method="GET",status="200"} 5
...
```

## Step 7: View Logs

```bash
# Get pod name
POD_NAME=$(oc get pods -l app=prometheus-demo-app -o jsonpath='{.items[0].metadata.name}')

# View logs
oc logs $POD_NAME

# Follow logs
oc logs -f $POD_NAME
```

## Step 8: Access the OpenShift Console

1. Visit: https://console-openshift-console.apps.sandbox.x8i5.p1.openshiftapps.com/
2. Login with your Red Hat Developer account
3. Navigate to your project (prometheus-demo)
4. Click on the Route to access your application
5. View logs and metrics from the console

## Prometheus Integration

### Verify Service Annotations

```bash
oc get svc prometheus-demo-app -o yaml
```

You should see:
```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/path: "/metrics"
    prometheus.io/port: "8080"
```

### Using ServiceMonitor (Optional)

If you have the Prometheus Operator installed:

```bash
oc apply -f servicemonitor.yaml
```

This creates a ServiceMonitor that the Prometheus Operator will use to discover and scrape your application.

## Troubleshooting

### Build Fails

If the build fails, check the build logs:
```bash
oc get builds
oc logs build/<build-name>
```

### Pod Not Starting

Check pod status and events:
```bash
oc get pods
oc describe pod <pod-name>
oc logs <pod-name>
```

### Image Pull Errors

If you see image pull errors:
```bash
# Check available image streams
oc get is

# Rebuild the image
oc start-build prometheus-demo-app --from-dir=. --follow
```

### Metrics Not Being Scraped

1. Verify the service annotations are correct
2. Check if Prometheus is running in your cluster:
```bash
oc get pods -n openshift-monitoring
```

3. Test the metrics endpoint directly:
```bash
oc port-forward svc/prometheus-demo-app 8080:8080
curl http://localhost:8080/metrics
```

### Permission Errors

The application runs as non-root (UID 65534). If you encounter permission errors:

```bash
# Check security context constraints
oc describe scc restricted

# Verify pod security context
oc get pod <pod-name> -o yaml | grep -A 10 securityContext
```

## Resource Limits

The application is configured with minimal resource usage suitable for Developer Sandbox:

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

If you need to adjust these limits:

```bash
oc set resources deployment/prometheus-demo-app --limits=cpu=500m,memory=256Mi --requests=cpu=200m,memory=128Mi
```

## Scaling

To scale the application:

```bash
# Scale to 3 replicas
oc scale deployment prometheus-demo-app --replicas=3

# Check scaling status
oc get pods
```

## Cleanup

To remove all resources:

```bash
# Using individual commands
oc delete route prometheus-demo-app
oc delete service prometheus-demo-app
oc delete deployment prometheus-demo-app
oc delete buildconfig prometheus-demo-app
oc delete imagestream prometheus-demo-app

# Or using the Makefile
make undeploy

# Or delete the entire project
oc delete project prometheus-demo
```

## Next Steps

1. **Monitor the Application**: Use the OpenShift monitoring stack or external Prometheus
2. **Create Grafana Dashboard**: Import the provided `grafana-dashboard.json`
3. **Set Up Alerts**: Configure Prometheus alerts based on the metrics
4. **Add More Metrics**: Extend the application with custom business metrics
5. **Integrate with OpenTelemetry**: Add OpenTelemetry support for distributed tracing

## Additional Resources

- [OpenShift Documentation](https://docs.openshift.com/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go Prometheus Client](https://github.com/prometheus/client_golang)