# tfdiff

A tool to compare differences between Terraform root modules. It analyzes module calls, outputs, resources, data sources, and variables to help identify changes between different versions or configurations of Terraform modules.

Features:
- **Attribute-level diff**: Shows only changed attributes for modified resources/modules
- **Multi-line formatting**: Clean, readable output similar to git diff
- **Configurable comparison**: Control what to compare and how detailed to be
- **HCL parsing**: Direct parsing of Terraform files for accurate attribute extraction

## Install

```bash
go install github.com/takaishi/tfdiff/cmd/tfdiff
```

## Usage

### Basic Comparison

Compare two Terraform modules:

```bash
tfdiff /path/to/module1 /path/to/module2
```

### Comparison Levels

Control what elements to compare using the `-l` flag:

```bash
# Compare only module calls and outputs
tfdiff module1 module2 -l module_calls,outputs

# Compare everything
tfdiff module1 module2 -l all

# Available levels: module_calls, outputs, resources, data_sources, variables, all
```

### Output Formats

Choose between text (default) and JSON output:

```bash
# Human-readable text output (default)
tfdiff module1 module2

# JSON output for programmatic use
tfdiff module1 module2 -o json
```

### Configuration Options

```bash
# Ignore argument differences (default: true)
tfdiff module1 module2 --ignore-args=false
```

## Example Output

Unified diff format showing attribute-level changes:

```diff
--- ./examples/basic/left
+++ ./examples/basic/right
+module "ecs_cluster" {
+  source  = "terraform-aws-modules/ecs/aws"
+  version = "~> 4.0"
+}
 module "rds" {
-  version = "~> 5.0"
+  version = "~> 6.0"
 }
 module "vpc" {
-  version = "~> 3.0"
+  version = "~> 4.0"
 }
-module "s3_bucket" {
-  source  = "terraform-aws-modules/s3-bucket/aws"
-  version = "~> 3.0"
-}
+output "alb_dns_name" {
+  description = "DNS name of the load balancer"
+  sensitive = false
+}
+resource "aws_cloudwatch_log_group" "app_logs" {
+}
-resource "aws_s3_bucket" "logs" {
-}
```

### Argument Comparison

Control whether to compare configuration arguments:

```bash
# Show argument differences (default: ignore arguments)
tfdiff module1 module2 --ignore-args=false

# Ignore argument differences (default behavior)
tfdiff module1 module2 --ignore-args=true
```

With `--ignore-args=false`, attribute-level differences are shown:

```diff
 module "rds" {
-  version = "~> 5.0"
+  version = "~> 6.0"
-  allocated_storage = "20"
+  allocated_storage = "50"
-  engine = "mysql"
+  engine = "postgresql"
-  engine_version = "8.0"
+  engine_version = "14.9"
-  instance_class = "db.t3.micro"
+  instance_class = "db.t3.small"
 }
```
