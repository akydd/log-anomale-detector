resource "aws_lambda_function" "log_generator" {
  function_name    = "log-generator"
  role             = aws_iam_role.log_generator.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  architectures    = ["arm64"]
  filename         = "${path.module}/dist/log-generator.zip"
  source_code_hash = filebase64sha256("${path.module}/dist/log-generator.zip")

  logging_config {
    log_format            = "JSON" # Recommended for easier querying
    application_log_level = "INFO"
    system_log_level      = "WARN"
    log_group             = aws_cloudwatch_log_group.shared_logging.name
  }
  depends_on = [aws_cloudwatch_log_group.shared_logging]
}

resource "aws_lambda_function" "detector" {
  function_name    = "detector"
  role             = aws_iam_role.detector.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  architectures    = ["arm64"]
  filename         = "${path.module}/dist/detector.zip"
  source_code_hash = filebase64sha256("${path.module}/dist/detector.zip")
  timeout          = 30

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
  timeout          = 180

  logging_config {
    log_format            = "JSON" # Recommended for easier querying
    application_log_level = "INFO"
    system_log_level      = "WARN"
    log_group             = aws_cloudwatch_log_group.shared_logging.name
  }

  depends_on = [aws_cloudwatch_log_group.shared_logging]
}
