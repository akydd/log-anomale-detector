# resource "aws_cloudwatch_log_group" "log_generator" {
#   name              = "/aws/lambda/log-generator"
#   retention_in_days = 7
# }

# resource "aws_cloudwatch_log_group" "chaos_injector" {
#   name              = "/aws/lambda/chaos-injector"
#   retention_in_days = 7
# }

resource "aws_cloudwatch_log_group" "shared_logging" {
  name              = "/aws/lambda/shared-logging"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "detector" {
  name              = "/aws/lambda/detector"
  retention_in_days = 7
}

resource "aws_lambda_permission" "detector" {
  statement_id  = "AllowCloudWatchLogsInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.detector.function_name
  principal     = "logs.amazonaws.com"
  source_arn    = "${aws_cloudwatch_log_group.shared_logging.arn}:*"
}

resource "aws_cloudwatch_log_subscription_filter" "detector" {
  name            = "detector-subscription"
  log_group_name  = aws_cloudwatch_log_group.shared_logging.name
  filter_pattern  = ""
  destination_arn = aws_lambda_function.detector.arn

  depends_on = [aws_lambda_permission.detector]
}
