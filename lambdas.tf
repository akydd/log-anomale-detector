resource "aws_lambda_function" "log_generator" {
  function_name    = "log-generator"
  role             = aws_iam_role.log_generator.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  architectures    = ["arm64"]
  filename         = "${path.module}/dist/log-generator.zip"
  source_code_hash = filebase64sha256("${path.module}/dist/log-generator.zip")

  depends_on = [aws_cloudwatch_log_group.log_generator]
}

resource "aws_lambda_function" "detector" {
  function_name    = "detector"
  role             = aws_iam_role.detector.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  architectures    = ["arm64"]
  filename         = "${path.module}/dist/detector.zip"
  source_code_hash = filebase64sha256("${path.module}/dist/detector.zip")

  environment {
    variables = {
      TABLE_NAME     = var.db_table
      SNS_TOPIC_ARN  = aws_sns_topic.alerts.arn
      AWS_ACCOUNT_ID = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_cloudwatch_log_group.detector]
}

resource "aws_lambda_function" "chaos_injector" {
  function_name    = "chaos-injector"
  role             = aws_iam_role.chaos_injector.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  architectures    = ["arm64"]
  filename         = "${path.module}/dist/chaos-injector.zip"
  source_code_hash = filebase64sha256("${path.module}/dist/chaos-injector.zip")

  depends_on = [aws_cloudwatch_log_group.chaos_injector]
}
