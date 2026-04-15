resource "aws_sns_topic" "alerts" {
  name = "log-anomaly-detector"
  tags = { Name = "log-anomaly-detector" }
}

resource "aws_sns_topic_subscription" "alerts_email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alerts_email
}
