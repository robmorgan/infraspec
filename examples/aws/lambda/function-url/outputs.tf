output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.main.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.main.arn
}

output "function_url" {
  description = "URL of the Lambda function"
  value       = aws_lambda_function_url.main.function_url
}

output "function_url_id" {
  description = "ID of the Lambda function URL"
  value       = aws_lambda_function_url.main.url_id
}

output "authorization_type" {
  description = "Authorization type of the function URL"
  value       = aws_lambda_function_url.main.authorization_type
}
