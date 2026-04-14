resource "aws_cloudwatch_event_rule" "log_generator" {
  name                = "log-generator-schedule"
  schedule_expression = "rate(1 minute)"
}

resource "aws_cloudwatch_event_target" "log_generator" {
  rule = aws_cloudwatch_event_rule.log_generator.name
  arn  = aws_lambda_function.log_generator.arn
}

resource "aws_lambda_permission" "log_generator" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.log_generator.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.log_generator.arn
}
