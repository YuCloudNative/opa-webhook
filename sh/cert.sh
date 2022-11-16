#!/bin/bash

SERVICE_NAME=opa-sidecar-webhook
NAMESPACE=opa-istio

cfssl gencert -initca ./tls/ca-csr.json | cfssljson -bare /tmp/ca

cfssl gencert \
  -ca=/tmp/ca.pem \
  -ca-key=/tmp/ca-key.pem \
  -config=./tls/ca-config.json \
  -hostname=${SERVICE_NAME},${SERVICE_NAME}.${NAMESPACE},${SERVICE_NAME}.${NAMESPACE}.svc,${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local,localhost,127.0.0.1 \
  -profile=default \
  ./tls/ca-csr.json | cfssljson -bare /tmp/opa-webhook

cp /tmp/opa-webhook.pem ./ssl/cert.pem
cp /tmp/opa-webhook-key.pem ./ssl/key.pem
cp /tmp/ca.pem ./ssl/ca.pem

cat ./ssl/ca.pem | base64 > ./ssl/ca-base64
cat ./ssl/key.pem | base64 > ./ssl/key-base64
cat ./ssl/cert.pem | base64 > ./ssl/cert-base64
