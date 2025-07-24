# tfdiff Examples

This directory contains example Terraform configurations to demonstrate tfdiff functionality.

## Basic Example

The `basic` directory contains two similar Terraform configurations with various differences:

### Left Module (Original)
- EC2 instance: `t2.micro` in `production` environment
- S3 bucket for logs
- VPC with NAT and VPN gateway enabled
- Security group with HTTP/HTTPS access only
- **Modules**: VPC (v3.0), RDS MySQL (v5.0), S3 bucket (v3.0)

### Right Module (Updated)
- EC2 instance: `t3.small` in `staging` environment
- CloudWatch log group instead of S3 bucket
- VPC with VPN gateway disabled, additional private subnet
- Security group with SSH access added
- New Application Load Balancer resource
- **Modules**: VPC (v4.0), RDS PostgreSQL (v6.0), ECS cluster (v4.0)

## Usage Examples

### 1. Basic Comparison (Default Output)
```bash
./tfdiff examples/basic/left examples/basic/right
```

### 2. JSON Output
```bash
./tfdiff examples/basic/left examples/basic/right --output json
```

### 3. Compare Only Resources
```bash
./tfdiff examples/basic/left examples/basic/right --level resources
```

### 4. Compare Only Module Calls
```bash
./tfdiff examples/basic/left examples/basic/right --level module_calls
```

### 5. Compare Only Variables and Outputs
```bash
./tfdiff examples/basic/left examples/basic/right --level variables,outputs
```

### 6. Compare Everything
```bash
./tfdiff examples/basic/left examples/basic/right --level all
```

### 7. Include Description Differences
```bash
./tfdiff examples/basic/left examples/basic/right --ignore-descriptions=false
```

### 8. Show Argument Differences
```bash
./tfdiff examples/basic/left examples/basic/right --ignore-args=false
```

### 9. Show Position Information
```bash
./tfdiff examples/basic/left examples/basic/right --ignore-positions=false
```

## Expected Differences

When comparing the basic example modules, you should see:

**Resources:**
- ➖ `aws_s3_bucket.logs` removed
- ➕ `aws_cloudwatch_log_group.app_logs` added
- ➕ `aws_lb.main` added

**Variables:**
- ➖ `bucket_name` removed
- ➕ `log_retention_days` added
- ➕ `enable_alb` added

**Outputs:**
- ➖ `bucket_name` removed
- ➕ `log_group_name` added
- ➕ `alb_dns_name` added
- ➕ `alb_zone_id` added

**Module Calls:**
- 📝 `vpc` module version updated (v3.0 → v4.0)
- 📝 `rds` module version updated (v5.0 → v6.0) with significant configuration changes
- ➖ `s3_bucket` module removed
- ➕ `ecs_cluster` module added

The differences will vary depending on which comparison levels you choose and which ignore flags you set.