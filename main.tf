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
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogGroup"]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/shared-logging:*"
      }
    ]
  })
}

resource "aws_iam_role" "detector" {
  name = "detector-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "detector" {
  name = "detector-lambda-policy"

  role = aws_iam_role.detector.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogGroup"]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/detector:*"
      },
      {
        Effect   = "Allow"
        Action   = ["dynamodb:PutItem"]
        Resource = "arn:aws:dynamodb:ca-west-1:${data.aws_caller_identity.current.account_id}:table/${var.db_table}"
      },
      {
        Effect   = "Allow"
        Action   = ["sns:Publish"]
        Resource = aws_sns_topic.alerts.arn
      },
      {
        Effect = "Allow"
        Action = ["bedrock:InvokeModel"]
        Resource = [
          "arn:aws:bedrock:ca-west-1:${data.aws_caller_identity.current.account_id}:inference-profile/global.anthropic.claude-opus-4-5-20251101-v1:0",
          "arn:aws:bedrock:*::foundation-model/anthropic.claude-opus-4-5-20251101-v1:0",
        ]
      }
    ]
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
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogGroup"]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/shared-logging:*"
      }
    ]
  })
}

resource "aws_iam_role" "query" {
  name = "query-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "query" {
  name = "query-lambda-policy"

  role = aws_iam_role.query.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogGroup"]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:ca-west-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/query:*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:Query",
          "dynamodb:GetItem",
        ]
        Resource = "arn:aws:dynamodb:ca-west-1:${data.aws_caller_identity.current.account_id}:table/${var.db_table}"
      },
    ]
  })
}
