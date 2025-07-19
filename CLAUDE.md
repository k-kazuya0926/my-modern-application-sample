# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Testing and Local Development
- Test individual Go applications: `cd applications/<app-name> && go test ./...`
- Run Go mod tidy: `cd applications/<app-name> && go mod tidy`
- Build locally: `cd applications/<app-name> && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go`

### Docker Build Commands
- Build Lambda container: `docker build -f applications/shared/lambda/Dockerfile --build-arg FUNCTION_NAME=<app-name> -t <image-name> applications/`
- Test container locally: `docker run --rm -p 9000:8080 <image-name>`

### CI/CD Workflows
- Static analysis: `docker run --rm -v "$(pwd):$(pwd)" -w "$(pwd)" rhysd/actionlint:latest`
- Security scan: Trivy scanner runs automatically on PR/push to main
- Lambda builds: Each application has its own build workflow triggering on file changes

## Architecture Overview

### Core Structure
This is a **serverless microservices architecture** built with Go 1.24 and AWS Lambda. Each application in the `applications/` directory is an independent Lambda function with its own purpose and dependencies.

### Key Patterns

**Lambda Function Structure:**
- Each application follows the same pattern: `main.go`, `go.mod`, `go.sum`
- All use `github.com/aws/aws-lambda-go/lambda` for Lambda runtime
- AWS SDK v2 is used consistently across applications (`github.com/aws/aws-sdk-go-v2/`)
- Global client initialization in `main()` function for connection reuse
- Environment variables for configuration (ENV, bucket names, etc.)

**Docker Build Pattern:**
- Shared Dockerfile at `applications/shared/lambda/Dockerfile`
- Multi-stage build: golang:1.24-alpine for build, AWS Lambda base image for runtime
- Static linking with security flags: `CGO_ENABLED=0`, `-ldflags='-w -s -extldflags "-static"'`
- Uses `FUNCTION_NAME` build arg to specify which application to build

**CI/CD Pattern:**
- Each Lambda function has dedicated GitHub Actions workflow
- Path-based triggering (only builds when relevant files change)
- ECR for container image storage
- AWS OIDC for secure credential management
- Composite actions for reusable build/deploy logic

### Application Categories

**Simple Functions:**
- `hello-world`: Basic Lambda handler pattern
- `tmp`: Experimental function

**AWS Service Integration:**
- `read-and-write-s3`: S3 event processing, ZIP encryption
- `register-user`: API Gateway + DynamoDB + S3 + SES integration
- `feature-flags`: AWS AppConfig integration
- `auth-by-cognito`: Cognito JWT token validation

**Message Processing:**
- `send-emails-via-sqs/`: Complete email system with SQS queuing, SES sending, bounce handling
- `fan-out/`: SNS/SQS fan-out pattern with multiple consumers

**Orchestration:**
- `saga-orchestration/`: Step Functions-based distributed transaction pattern with compensating actions

### Key Implementation Patterns

**Error Handling:**
- Panic recovery with logging in Lambda handlers
- Graceful degradation (e.g., email sending failures don't break user registration)
- Structured error responses for API Gateway

**AWS SDK Usage:**
- V2 SDK with context-aware operations
- Service-specific clients initialized globally
- Proper attribute value handling for DynamoDB

**Environment Configuration:**
- Environment-specific resource naming: `my-modern-application-sample-{env}-{resource}`
- Required environment variables validated at startup

**Security:**
- Static binary compilation to reduce attack surface
- No hardcoded credentials
- Presigned URLs for secure S3 access

## Infrastructure Integration

This codebase is designed to work with infrastructure managed at: https://github.com/k-kazuya0926/my-modern-application-sample-infra

AWS Services used:
- **API Gateway**: HTTP APIs for web endpoints
- **Lambda**: Serverless compute with container images
- **DynamoDB**: NoSQL database with sequence tables for ID generation
- **S3**: Object storage with event triggers and presigned URLs
- **SES**: Email sending with bounce handling
- **SQS/SNS**: Message queuing and pub/sub
- **Step Functions**: Workflow orchestration for saga pattern
- **AppConfig**: Feature flag management
- **Cognito**: User authentication
- **ECR**: Container image registry
- **X-Ray**: Distributed tracing

## Development Notes

### Adding New Lambda Functions
1. Create new directory under `applications/`
2. Add `go.mod`, `go.sum`, and `main.go` following existing patterns
3. Create corresponding GitHub Actions workflow following naming convention
4. Use shared Dockerfile with appropriate `FUNCTION_NAME` build arg

### Common Dependencies
- AWS Lambda Go runtime: `github.com/aws/aws-lambda-go`
- AWS SDK v2: `github.com/aws/aws-sdk-go-v2/`
- Specific services imported as needed (dynamodb, s3, ses, etc.)

### Environment Variables
All Lambda functions expect:
- `ENV`: Environment name (prod, dev, etc.)
- Service-specific variables (bucket names, table names, etc.)