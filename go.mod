module github.com/FreePeak/infra-mcp-server

go 1.23

toolchain go1.24.1

require (
	github.com/FreePeak/cortex v1.0.5
	github.com/aws/aws-sdk-go-v2 v1.39.6
	github.com/aws/aws-sdk-go-v2/config v1.31.20
	github.com/aws/aws-sdk-go-v2/credentials v1.18.24
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.52.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.58.9
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.267.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.67.4
	github.com/aws/aws-sdk-go-v2/service/lambda v1.81.3
	github.com/aws/aws-sdk-go-v2/service/rds v1.108.9
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.39.13
	github.com/go-sql-driver/mysql v1.9.1
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	go.uber.org/zap v1.27.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.40.2 // indirect
	github.com/aws/smithy-go v1.23.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
