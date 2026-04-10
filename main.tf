provider "aws" {
  region = "ca-west-1"
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "log_generator" {
  name = "log-generator-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "log_generator" {
  name = "log-generator-lambda-policy"

  role = aws_iam_role.log_generator.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
      Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/log-generator:*"
    }]
  })
}

resource "aws_iam_role" "chaos_injector" {
  name = "chaos-injector-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "chaos_injector" {
  name = "chaos-injector-lambda-policy"

  role = aws_iam_role.chaos_injector.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
      Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/chaos-injector:*"
    }]
  })
}
