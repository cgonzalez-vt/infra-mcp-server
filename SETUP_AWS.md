# AWS Integration Setup Guide

## Quick Start for Developers

This guide helps you set up AWS integration in the infra-mcp-server project.

## Prerequisites

- AWS account with read-only access
- IAM user with appropriate permissions (see IAM Policy below)
- Access key ID and secret access key

## Step 1: Get AWS Credentials

1. Log into AWS Console
2. Navigate to IAM → Users
3. Create a new user or select existing user
4. Go to "Security credentials" tab
5. Click "Create access key"
6. Choose "Application running outside AWS"
7. Copy the Access Key ID and Secret Access Key

## Step 2: Configure the MCP Server

1. Open `bin/config.json`
2. Locate the `aws_profiles` section
3. Replace placeholder values with your actual credentials:

```json
{
  "aws_profiles": [
    {
      "id": "staging",
      "access_key_id": "YOUR_ACTUAL_ACCESS_KEY_ID",
      "secret_access_key": "YOUR_ACTUAL_SECRET_ACCESS_KEY",
      "region": "us-east-1",
      "project": "infrastructure",
      "environment": "staging",
      "description": "Staging environment read-only access",
      "tags": ["staging", "infrastructure", "read-only"]
    }
  ]
}
```

## Step 3: Secure Your Configuration

**IMPORTANT**: Never commit actual credentials to git!

### Option 1: Use .gitignore (Recommended)

Add to `.gitignore`:

```
bin/config.json
```

Then create a template file `bin/config.example.json` with placeholder values.

### Option 2: Use Environment Variables

Alternatively, you can store credentials in environment variables and modify the code to read from them.

### Option 3: Use AWS Secrets Manager

For production deployments, fetch credentials from AWS Secrets Manager at runtime.

## Step 4: Verify Setup

1. Build the project:

```bash
make build
```

2. Start the server:

```bash
./bin/server -t stdio -c bin/config.json
```

3. Check logs for AWS profile initialization:

```
Initialized AWS profile: staging (Staging environment read-only access)
```

## Required IAM Policy

Attach this policy to your IAM user for read-only access:

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

## Multiple Profiles

You can configure multiple AWS profiles for different environments:

```json
{
  "aws_profiles": [
    {
      "id": "staging",
      "access_key_id": "STAGING_KEY",
      "secret_access_key": "STAGING_SECRET",
      "region": "us-east-1",
      "project": "infrastructure",
      "environment": "staging",
      "description": "Staging environment",
      "tags": ["staging"]
    },
    {
      "id": "production",
      "access_key_id": "PRODUCTION_KEY",
      "secret_access_key": "PRODUCTION_SECRET",
      "region": "us-east-1",
      "project": "infrastructure",
      "environment": "production",
      "description": "Production environment",
      "tags": ["production"]
    }
  ]
}
```

Each profile will create separate tools:

- `aws_logs_list_staging`
- `aws_logs_list_production`
- etc.

## Troubleshooting

### "access_key_id cannot be empty"

- Ensure you've replaced the placeholder values in config.json
- Check that the JSON is valid (no trailing commas, proper quotes)

### "AccessDenied" errors

- Verify your IAM user has the required permissions
- Check the IAM policy is attached to your user
- Ensure the access key is active (not disabled)

### "InvalidClientTokenId"

- The access key ID is incorrect
- The access key may have been deleted
- Check for typos in the access key ID

### "SignatureDoesNotMatch"

- The secret access key is incorrect
- Check for extra spaces or newlines in the secret
- Ensure you copied the complete secret key

## Security Checklist

- [ ] IAM user has ONLY read-only permissions
- [ ] Access keys are not committed to git
- [ ] config.json is in .gitignore
- [ ] Credentials are rotated regularly (every 90 days)
- [ ] Separate credentials for staging and production
- [ ] MFA enabled on AWS account
- [ ] CloudTrail logging enabled for audit

## Getting Help

If you encounter issues:

1. Check the server logs in `logs/stdio-server.log`
2. Verify AWS credentials with AWS CLI: `aws sts get-caller-identity`
3. Test permissions with AWS CLI: `aws logs describe-log-groups --max-items 1`
4. Review the AWS_INTEGRATION.md for detailed tool documentation
