# AWS Infrastructure Integration

This document describes the AWS infrastructure integration in the Infrastructure MCP Server, which allows agents to inspect and monitor AWS resources without using CLI commands explicitly.

## Overview

The AWS integration provides read-only access to various AWS services through MCP tools. Agents can query CloudWatch logs, inspect ECS clusters, check RDS status, and access other AWS resources using the configured AWS profiles.

## Supported AWS Services

- **CloudWatch Logs**: Query and tail log groups and streams
- **ECS (Elastic Container Service)**: List and describe clusters, services, and tasks
- **RDS (Relational Database Service)**: List and describe database instances
- **EC2 (Elastic Compute Cloud)**: List EC2 instances
- **Lambda**: List Lambda functions
- **Secrets Manager**: List secrets (metadata only, not values)

## Configuration

### Self-Contained Configuration

AWS credentials are stored directly in the `config.json` file, making it easy for developers to check out the project and start using it without additional setup.

**Important**: Never commit actual AWS credentials to version control. Use placeholder values and have developers replace them with their actual credentials locally.

### MCP Server Configuration

Add AWS profiles to your `config.json`:

```json
{
  "connections": [...],
  "aws_profiles": [
    {
      "id": "staging",
      "access_key_id": "AKIAIOSFODNN7EXAMPLE",
      "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
      "region": "us-east-1",
      "project": "infrastructure",
      "environment": "staging",
      "description": "Staging environment read-only access",
      "tags": ["staging", "infrastructure", "read-only"]
    }
  ]
}
```

### Configuration Fields

- `id` (required): Unique identifier for this profile (used in tool names)
- `access_key_id` (required): AWS access key ID
- `secret_access_key` (required): AWS secret access key
- `region` (optional): AWS region (defaults to us-east-1)
- `project` (optional): Project name for organization
- `environment` (optional): Environment name (staging, production, etc.)
- `description` (optional): Human-readable description
- `tags` (optional): Array of tags for categorization

### Security Best Practices

1. **Never commit credentials**: Add `config.json` to `.gitignore` or use placeholder values
2. **Use read-only IAM users**: Create dedicated IAM users with read-only permissions
3. **Rotate credentials regularly**: Update credentials periodically
4. **Use environment-specific credentials**: Separate credentials for staging and production
5. **Consider using AWS Secrets Manager**: For production deployments, consider fetching credentials from Secrets Manager

## Available Tools

All AWS tools follow the naming pattern: `aws_<service>_<action>_<profile_id>`

### CloudWatch Logs Tools

#### `aws_logs_list_<profile>`

List CloudWatch log groups.

**Parameters:**

- `prefix` (string, optional): Filter log groups by prefix
- `limit` (number, optional): Maximum number of log groups to return (default: 50)

**Example:**

```json
{
  "tool": "aws_logs_list_staging",
  "parameters": {
    "prefix": "/ecs/",
    "limit": 20
  }
}
```

#### `aws_logs_query_<profile>`

Query CloudWatch log events.

**Parameters:**

- `log_group` (string, required): Log group name
- `filter_pattern` (string, optional): CloudWatch filter pattern
- `start_time` (number, optional): Start time in milliseconds since epoch
- `end_time` (number, optional): End time in milliseconds since epoch
- `limit` (number, optional): Maximum number of events (default: 100)

**Example:**

```json
{
  "tool": "aws_logs_query_staging",
  "parameters": {
    "log_group": "/ecs/staging-payments-service",
    "filter_pattern": "ERROR",
    "limit": 50
  }
}
```

### ECS Tools

#### `aws_ecs_clusters_<profile>`

List all ECS clusters.

**Example:**

```json
{
  "tool": "aws_ecs_clusters_staging"
}
```

#### `aws_ecs_services_<profile>`

List services in an ECS cluster.

**Parameters:**

- `cluster_name` (string, required): Cluster name or ARN

**Example:**

```json
{
  "tool": "aws_ecs_services_staging",
  "parameters": {
    "cluster_name": "stg-payments-ecs-cluster"
  }
}
```

### RDS Tools

#### `aws_rds_list_<profile>`

List all RDS database instances.

**Example:**

```json
{
  "tool": "aws_rds_list_staging"
}
```

#### `aws_rds_describe_<profile>`

Get detailed information about an RDS instance.

**Parameters:**

- `identifier` (string, required): DB instance identifier

**Example:**

```json
{
  "tool": "aws_rds_describe_staging",
  "parameters": {
    "identifier": "stg-payment-gateway-instance1"
  }
}
```

### EC2 Tools

#### `aws_ec2_instances_<profile>`

List all EC2 instances.

**Example:**

```json
{
  "tool": "aws_ec2_instances_staging"
}
```

### Lambda Tools

#### `aws_lambda_list_<profile>`

List all Lambda functions.

**Example:**

```json
{
  "tool": "aws_lambda_list_staging"
}
```

### Secrets Manager Tools

#### `aws_secrets_list_<profile>`

List Secrets Manager secrets (metadata only, not secret values).

**Example:**

```json
{
  "tool": "aws_secrets_list_staging"
}
```

## Security Considerations

- **Read-Only Access**: All AWS tools are read-only by design. No write, delete, or modify operations are exposed.
- **IAM Permissions**: The AWS profile should have read-only permissions. Example IAM policy is provided below.
- **Credential Management**: AWS credentials are loaded from standard AWS configuration files (`~/.aws/credentials` and `~/.aws/config`).
- **Secret Values**: The Secrets Manager tool only lists secret metadata, not actual secret values.

### Example IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CloudWatchLogsReadOnly",
      "Effect": "Allow",
      "Action": [
        "logs:Describe*",
        "logs:Get*",
        "logs:List*",
        "logs:FilterLogEvents"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ECSReadOnly",
      "Effect": "Allow",
      "Action": ["ecs:Describe*", "ecs:List*"],
      "Resource": "*"
    },
    {
      "Sid": "RDSReadOnly",
      "Effect": "Allow",
      "Action": ["rds:Describe*", "rds:ListTagsForResource"],
      "Resource": "*"
    },
    {
      "Sid": "EC2ReadOnly",
      "Effect": "Allow",
      "Action": ["ec2:Describe*", "ec2:Get*"],
      "Resource": "*"
    },
    {
      "Sid": "LambdaReadOnly",
      "Effect": "Allow",
      "Action": ["lambda:Get*", "lambda:List*"],
      "Resource": "*"
    },
    {
      "Sid": "SecretsManagerReadOnly",
      "Effect": "Allow",
      "Action": ["secretsmanager:DescribeSecret", "secretsmanager:ListSecrets"],
      "Resource": "*"
    }
  ]
}
```

## Architecture

The AWS integration follows the same Clean Architecture pattern as the database tools:

```
pkg/aws/                    - AWS SDK clients and service wrappers
├── config.go              - AWS configuration management
├── clients.go             - Client manager for all AWS services
├── cloudwatch.go          - CloudWatch Logs operations
├── ecs.go                 - ECS operations
├── rds.go                 - RDS operations
├── ec2.go                 - EC2 operations
├── lambda.go              - Lambda operations
├── secrets.go             - Secrets Manager operations
└── cloudwatch_metrics.go  - CloudWatch Metrics operations

internal/delivery/mcp/
└── aws_manager.go         - MCP tool registration and handling
```

## Error Handling

All AWS operations include comprehensive error handling:

- AWS SDK errors are properly propagated
- Network timeouts are handled gracefully
- Permission errors provide helpful messages
- Rate limiting is respected

## Logging

All AWS operations are logged for audit and debugging purposes:

- Profile initialization
- Tool registration
- API calls (without sensitive data)
- Errors and warnings

## Future Enhancements

Potential future additions:

- CloudWatch Metrics querying
- S3 bucket listing and object inspection
- DynamoDB table inspection
- SNS/SQS queue monitoring
- Cost Explorer integration
- CloudFormation stack inspection

## Troubleshooting

### AWS Profile Not Found

**Error**: `profile fs-mcp-staging not found`

**Solution**: Ensure the profile exists in `~/.aws/config` and `~/.aws/credentials`

### Permission Denied

**Error**: `AccessDenied` or `UnauthorizedOperation`

**Solution**: Verify the IAM user/role has the required read-only permissions

### No Tools Registered

**Check**: Ensure `aws_profiles` is properly configured in `config.json`

### Connection Timeout

**Solution**: Check network connectivity to AWS services and verify security group rules

## Testing

To test the AWS integration:

1. Configure a test AWS profile with read-only permissions
2. Add the profile to `config.json`
3. Restart the MCP server
4. Verify tools are registered in the startup logs
5. Test each tool with sample queries

## Performance Considerations

- AWS API calls are made on-demand (no caching)
- Rate limits are respected automatically by the AWS SDK
- Large result sets may take time to retrieve
- Consider using filters and limits to reduce response sizes

