variable "db_table" {
  description = "DynamoDB table name"
  type        = string
}

variable "alerts_email" {
  description = "Email that receives alerts"
  type        = string
}

variable "model_id" {
  description = "Bedrock inference profile ID"
  type        = string
}