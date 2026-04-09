# Serverless Log Anomaly Detector with Chaos Injection

## Overview

A serverless AWS project that simulates realistic application logs, injects anomalous traffic patterns, detects anomalies in real time using Bedrock (Claude), and alerts via SNS. Built entirely in Go, deployed with CDK.

## Architecture

### Log Generation Layer

- **`log-generator` Lambda** — triggered by EventBridge every 1 minute, writes realistic HTTP access logs to CloudWatch (normal traffic: 200s, occasional 404s)
- **`chaos-injector` Lambda** — manually invoked, writes a burst of anomalous log entries for a configurable duration. Anomaly types:
  - 500 error spike
  - Repeated auth failures (simulated brute force)
  - Latency outliers

### Detection Pipeline

- **CloudWatch Logs Subscription Filter** → **Kinesis Data Stream** (real-time ingestion)
- **`anomaly-detector` Lambda** — triggered by Kinesis batches, sends log entries to Bedrock (Claude) with a prompt to flag and classify anomalies:
  - Error spike
  - Auth attack
  - Latency degradation
- Results written to **DynamoDB**: timestamp, anomaly type, severity, raw evidence, Bedrock explanation

### Alerting

- HIGH severity anomalies → **SNS** → email notification with Bedrock's plain-English incident summary

### Dashboard / API
- **API Gateway** + **`query` Lambda** → reads DynamoDB → returns anomaly history as JSON
- Optional: S3-hosted static HTML dashboard that polls the API and displays incidents

## Infrastructure

- **IaC:** AWS CDK (TypeScript)
- **CI/CD:** GitHub Actions
  - On PR: lint → unit tests → `cdk diff`
  - On merge to main: `cdk deploy`

## Language

All Lambdas written in **Go**:
- Fast cold starts (strong Lambda performance story)
- Single binary deploys
- AWS SDK for Go v2

## Demo Flow

1. Deploy the stack with `cdk deploy`
2. Show the dashboard — normal traffic, no anomalies
3. Invoke the chaos Lambda via CLI (or dashboard button)
4. Within ~60 seconds: anomaly appears in DynamoDB, SNS alert email arrives
5. Dashboard updates with classified incident and Bedrock's plain-English explanation

## AWS Services Used

| Service | Purpose |
|---|---|
| Lambda | All compute (generator, chaos, detector, query) |
| EventBridge | Schedules log-generator every 1 minute |
| CloudWatch Logs | Log ingestion and subscription filter |
| Kinesis Data Streams | Real-time log streaming to detector |
| Bedrock (Claude) | Anomaly classification and explanation |
| DynamoDB | Anomaly record storage |
| SNS | High-severity email alerting |
| API Gateway | REST API for dashboard |
| S3 | Static dashboard hosting (optional) |
| CDK | Infrastructure as code |
