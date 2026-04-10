resource "aws_cloudwatch_log_group" "log_generator" {
  name              = "/aws/lambda/log-generator"
  retention_in_days = 7
}


resource "aws_cloudwatch_log_group" "chaos_injector" {
  name              = "/aws/lambda/chaos-injector"
  retention_in_days = 7
}
