resource "aws_api_gateway_rest_api" "main" {
  name = "log-anomaly-detector"
}

resource "aws_api_gateway_resource" "chaos" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_rest_api.main.root_resource_id
  path_part   = "chaos"
}

resource "aws_api_gateway_method" "chaos_post" {
  rest_api_id      = aws_api_gateway_rest_api.main.id
  resource_id      = aws_api_gateway_resource.chaos.id
  http_method      = "POST"
  authorization    = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "chaos_post" {
  rest_api_id             = aws_api_gateway_rest_api.main.id
  resource_id             = aws_api_gateway_resource.chaos.id
  http_method             = aws_api_gateway_method.chaos_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.chaos_injector.invoke_arn
}

resource "aws_api_gateway_deployment" "main" {
  rest_api_id = aws_api_gateway_rest_api.main.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.chaos,
      aws_api_gateway_method.chaos_post,
      aws_api_gateway_integration.chaos_post,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [aws_api_gateway_integration.chaos_post]
}

resource "aws_api_gateway_stage" "prod" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  deployment_id = aws_api_gateway_deployment.main.id
  stage_name    = "prod"
}

resource "aws_api_gateway_api_key" "main" {
  name = "log-anomaly-detector-key"
}

resource "aws_api_gateway_usage_plan" "main" {
  name = "log-anomaly-detector-plan"

  api_stages {
    api_id = aws_api_gateway_rest_api.main.id
    stage  = aws_api_gateway_stage.prod.stage_name
  }
}

resource "aws_api_gateway_usage_plan_key" "main" {
  key_id        = aws_api_gateway_api_key.main.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.main.id
}

resource "aws_lambda_permission" "chaos_injector" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.chaos_injector.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.main.execution_arn}/*/*"
}

output "chaos_endpoint" {
  value = "${aws_api_gateway_stage.prod.invoke_url}/chaos"
}

output "api_key" {
  value     = aws_api_gateway_api_key.main.value
  sensitive = true
}
