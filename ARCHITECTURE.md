# Serverless Log Anomaly Detector with Chaos Injection

## Overview

A serverless AWS project that simulates realistic application logs, injects anomalous traffic patterns, detects anomalies in real time using Bedrock (Claude), and alerts via SNS. Built entirely in Go, deployed with Terraform. No VPC — security is enforced through least-privilege IAM roles and resource-based policies.

## Architecture

### Log Generation Layer

- **`log-generator` Lambda** — triggered by EventBridge every 1 minute, writes realistic HTTP access logs to CloudWatch (normal traffic: 200s, occasional 404s)
- **`chaos-injector` Lambda** — invoked via API Gateway, writes a burst of anomalous log entries for a configurable duration. Anomaly types:
  - 500 error spike
  - Repeated auth failures (simulated brute force)
  - Latency outliers

### Detection Pipeline

- **CloudWatch Logs Subscription Filter** → **`anomaly-detector` Lambda** (real-time, event-driven)
- CloudWatch pushes compressed, base64-encoded log batches directly to the Lambda as they arrive
- **`anomaly-detector` Lambda** decodes the batch, sends log entries to Bedrock (Claude) with a prompt to flag and classify anomalies:
  - Error spike
  - Auth attack
  - Latency degradation
- Results written to **DynamoDB**: timestamp, anomaly type, severity, raw evidence, Bedrock explanation

### Alerting

- HIGH severity anomalies → **SNS** → email notification with Bedrock's plain-English incident summary

### Dashboard / API

- **API Gateway** (API key auth) + **`query` Lambda** → reads DynamoDB → returns anomaly history as JSON
- Optional: S3-hosted static HTML dashboard that polls the API and displays incidents

## Security

Security is enforced through IAM identity and resource policies — not network topology. Each Lambda has a dedicated execution role scoped to the minimum actions it requires.

### IAM Execution Roles (least-privilege)

| Lambda | Allowed Actions | Scope |
|---|---|---|
| `log-generator` | `logs:PutLogEvents`, `logs:CreateLogStream` | Its own log group only |
| `chaos-injector` | `logs:PutLogEvents`, `logs:CreateLogStream` | Its own log group only |
| `anomaly-detector` | `bedrock:InvokeModel` | One model ARN (Claude) |
| `anomaly-detector` | `dynamodb:PutItem` | One table |
| `anomaly-detector` | `sns:Publish` | One topic ARN |
| `query` | `dynamodb:Query`, `dynamodb:GetItem` | One table |

No Lambda role grants cross-function access. A compromised `log-generator` cannot read DynamoDB or invoke Bedrock.

### Resource-Based Policies

Resources are locked from the receiving side as well:

- **anomaly-detector Lambda** — accepts invocations only from the CloudWatch Logs service principal (`logs.amazonaws.com`) for the log group's subscription filter
- **DynamoDB table** — accepts calls only from the `anomaly-detector` and `query` Lambda execution roles
- **SNS topic** — accepts publishes only from the `anomaly-detector` Lambda execution role
- **API Gateway** — protected by an API key; only requests with a valid key reach the `query` Lambda

### Encryption

- DynamoDB and CloudWatch Logs are encrypted at rest by default
- SNS messages encrypted in transit via TLS
- All Lambda-to-AWS-service calls use HTTPS (AWS SDK default)

### Why No VPC

None of the services in this stack require VPC placement — all communication is between Lambda and AWS-managed services over AWS public endpoints. Adding a VPC would require NAT Gateways (~$32–65/month per deployment) and add operational complexity with no security benefit, since IAM policies already enforce strict per-resource access controls. Network isolation does not substitute for identity-based access control in a serverless architecture.

## Infrastructure

- **IaC:** Terraform
- **CI/CD:** GitHub Actions
  - On PR: lint → unit tests → `terraform plan`
  - On merge to main: `terraform apply`

## Language

All Lambdas written in **Go**:
- Fast cold starts (strong Lambda performance story)
- Single binary deploys
- AWS SDK for Go v2

## Demo Flow

1. Deploy the stack with `terraform apply`
2. Show the dashboard — normal traffic, no anomalies
3. Invoke the chaos Lambda via the API (`POST /chaos`)
4. Within ~60 seconds: anomaly appears in DynamoDB, SNS alert email arrives
5. Dashboard updates with classified incident and Bedrock's plain-English explanation

## AWS Services Used

| Service | Purpose |
|---|---|
| Lambda | All compute (generator, chaos, detector, query) |
| EventBridge | Schedules log-generator every 1 minute |
| CloudWatch Logs | Log ingestion and subscription filter → anomaly-detector |
| Bedrock (Claude) | Anomaly classification and explanation |
| DynamoDB | Anomaly record storage |
| SNS | High-severity email alerting |
| API Gateway | REST API for chaos injection and dashboard (API key auth) |
| S3 | Static dashboard hosting (optional) |
| IAM | Least-privilege execution roles and resource policies |
| Terraform | Infrastructure as code |
