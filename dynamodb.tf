resource "aws_dynamodb_table" "classified_logs" {
  name         = var.db_table
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "timestamp"

  attribute {
    name = "timestamp"
    type = "S"
  }
}
