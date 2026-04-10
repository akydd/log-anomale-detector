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
