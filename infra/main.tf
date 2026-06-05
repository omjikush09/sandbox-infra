terraform {
  required_version = "~>  1.15.4"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.AWS_REGION
}

module "ec2" {
  source = "./modules/ec2"
}

variable "AWS_REGION" {
  type = string
}
