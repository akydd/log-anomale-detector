resource "aws_s3_bucket" "dashboard" {
  bucket_prefix = "log-anomaly-dashboard-"
}

resource "aws_s3_bucket_public_access_block" "dashboard" {
  bucket = aws_s3_bucket.dashboard.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_website_configuration" "dashboard" {
  bucket = aws_s3_bucket.dashboard.id

  index_document {
    suffix = "index.html"
  }
}

resource "aws_s3_bucket_policy" "dashboard" {
  bucket = aws_s3_bucket.dashboard.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.dashboard.arn}/*"
    }]
  })

  depends_on = [aws_s3_bucket_public_access_block.dashboard]
}

resource "aws_s3_object" "dashboard" {
  bucket       = aws_s3_bucket.dashboard.id
  key          = "index.html"
  content_type = "text/html"
  content = templatefile("${path.module}/dashboard/index.html", {
    query_endpoint = "${aws_api_gateway_stage.prod.invoke_url}/query"
    api_key        = aws_api_gateway_api_key.main.value
  })
  etag = md5(templatefile("${path.module}/dashboard/index.html", {
    query_endpoint = "${aws_api_gateway_stage.prod.invoke_url}/query"
    api_key        = aws_api_gateway_api_key.main.value
  }))
}

output "dashboard_url" {
  value = "http://${aws_s3_bucket_website_configuration.dashboard.website_endpoint}"
}
