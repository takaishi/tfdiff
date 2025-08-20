# Basic Example - JSON Object Comparison

This directory contains Terraform configurations that demonstrate tfdiff's basic functionality, particularly how it handles JSON object comparison when key ordering differs.

## JSON Object Ordering Differences

### aws_instance.web tags
- **left**: `Name`, `Environment`, `Team`, `Project` ordering
- **right**: `Project`, `Environment`, `Name`, `Team`, `Owner` ordering
- Common tags (Name, Team, Project) are considered equal despite different ordering
- Environment changed from "production" → "staging" (actual difference)
- Owner tag added only in right (actual difference)


## Usage Examples

```bash
# Basic comparison
tfdiff left right

# Resources only comparison
tfdiff left right --level resources

# Include argument differences
tfdiff left right --ignore-args=false
```

## Expected Results

JSON object ordering differences are not detected as changes, only actual content differences are shown:

- aws_instance.web: 
  - instance_type change (t2.micro → t3.small)
  - Environment tag change (production → staging)
  - Owner tag addition
- aws_security_group.web_sg: Modified (ingress rules changed)
- aws_s3_bucket.logs removal
- aws_cloudwatch_log_group.app_logs addition
- Other module and resource changes