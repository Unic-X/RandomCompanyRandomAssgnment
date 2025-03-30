#!/bin/bash

echo "Building Docker image..."
docker build -t slow-server:latest ./server

echo "Deploying to Kubernetes..."
kubectl apply -k k8s/

echo "Waiting for deployments to be ready..."
kubectl -n slow-server wait --for=condition=available --timeout=300s deployment/slow-server
kubectl -n slow-server wait --for=condition=available --timeout=300s deployment/prometheus
kubectl -n slow-server wait --for=condition=available --timeout=300s deployment/grafana

echo ""
echo "Deployment complete! Service information:"
echo ""
echo "Slow Server:"
kubectl -n slow-server get svc slow-server
echo ""
echo "Grafana Dashboard:"
kubectl -n slow-server get svc grafana
echo ""
echo "Prometheus Status:"
kubectl -n slow-server get svc prometheus
echo ""
echo "Use the EXTERNAL-IP values to access the services."
echo "Grafana login: admin / admin"
