# deploy.ps1
Write-Host "Building Docker images..." -ForegroundColor Green
docker build -t weather-collector:latest ./collector
docker build -t weather-analyzer:latest ./analyzer

Write-Host "Loading images into minikube..." -ForegroundColor Green
minikube image load weather-collector:latest
minikube image load weather-analyzer:latest

Write-Host "Applying Kubernetes manifests..." -ForegroundColor Green
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/collector-deployment.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/analyzer-deployment.yaml

Write-Host "Checking status..." -ForegroundColor Green
kubectl get pods -n data-pipeline
kubectl get hpa -n data-pipeline

Write-Host "Dashboard URL:" -ForegroundColor Yellow
minikube service weather-analyzer-service -n data-pipeline --url