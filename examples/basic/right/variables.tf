variable "instance_type" {
  description = "Type of EC2 instance to launch"
  type        = string
  default     = "t3.small" # Changed default from t2.micro
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "staging" # Changed from production
}

# bucket_name variable removed

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# New variables added
variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 14
}

variable "enable_alb" {
  description = "Whether to create an Application Load Balancer"
  type        = bool
  default     = true
}