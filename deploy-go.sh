#!/bin/bash
# deploy-go.sh
LOG_LEVEL="INFO"

gcloud artifacts repositories create fin-tracker-go-api \
  --repository-format=docker \
  --location=us-central1 \
  --quiet 2>/dev/null || true

docker build -f Dockerfile.goapi \
  -t us-central1-docker.pkg.dev/project-10ae614e-b7c1-465b-93a/fin-tracker-go-api/fin-tracker-go-api . \
  && docker push us-central1-docker.pkg.dev/project-10ae614e-b7c1-465b-93a/fin-tracker-go-api/fin-tracker-go-api \
  && gcloud run deploy fin-tracker-go-api \
    --image us-central1-docker.pkg.dev/project-10ae614e-b7c1-465b-93a/fin-tracker-go-api/fin-tracker-go-api \
    --region us-central1 \
    --allow-unauthenticated \
    --set-env-vars MONGO_ATLAS_CONN_STR=$MONGO_ATLAS_CONN_STR \
    --set-env-vars LOG_LEVEL=$LOG_LEVEL \
    --set-env-vars LOG_FORMAT=json