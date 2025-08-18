# Web application infrastructure - Updated version
resource "aws_instance" "web" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t3.small" # Changed from t2.micro

  # JSON object with different key ordering, but Team and Project are the same
  tags = {
    Project     = "web-app"   # Different ordering
    Environment = "staging"   # Changed from production
    Name        = "WebServer" # Different ordering
    Team        = "backend"   # Different ordering
    Owner       = "DevOps"    # Added new tag
  }
}

resource "aws_security_group" "web_sg" {
  name_prefix = "web-sg"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Added SSH access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "WebSecurityGroup"
  }
}

# S3 bucket removed - replaced with CloudWatch logs
resource "aws_cloudwatch_log_group" "app_logs" {
  name              = "/aws/application/my-app"
  retention_in_days = 14

  tags = {
    Name        = "AppLogGroup"
    Environment = "staging"
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 4.0" # Updated version

  name = "my-vpc"
  cidr = "10.0.0.0/16"

  azs             = data.aws_availability_zones.available.names
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"] # Added third subnet
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = false # Changed from true

  tags = {
    Terraform   = "true"
    Environment = "staging" # Changed from production
  }
}

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.0" # Updated version

  identifier = "my-database-staging" # Changed name

  engine            = "postgresql"  # Changed from mysql
  engine_version    = "14.9"        # Updated version
  instance_class    = "db.t3.small" # Upgraded instance
  allocated_storage = 50            # Increased storage

  db_name  = "myapp_staging" # Changed name
  username = "dbadmin"       # Changed username

  vpc_security_group_ids = [aws_security_group.web_sg.id]

  backup_retention_period = 7 # Added backup retention

  tags = {
    Environment = "staging"
  }
}

# s3_bucket module removed

module "ecs_cluster" {
  source  = "terraform-aws-modules/ecs/aws"
  version = "~> 4.0"

  cluster_name = "my-app-cluster"

  cluster_configuration = {
    execute_command_configuration = {
      logging = "OVERRIDE"
      log_configuration = {
        cloud_watch_log_group_name = aws_cloudwatch_log_group.app_logs.name
      }
    }
  }

  fargate_capacity_providers = {
    FARGATE = {
      default_capacity_provider_strategy = {
        weight = 50
      }
    }
    FARGATE_SPOT = {
      default_capacity_provider_strategy = {
        weight = 50
      }
    }
  }

  tags = {
    Environment = "staging"
  }
}

# New ALB resource added
resource "aws_lb" "main" {
  name               = "main-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.web_sg.id]

  tags = {
    Name        = "MainALB"
    Environment = "staging"
  }
}
