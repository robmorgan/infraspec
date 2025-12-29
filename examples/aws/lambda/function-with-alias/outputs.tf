output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.main.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.main.arn
}

output "function_version" {
  description = "Published version of the Lambda function"
  value       = aws_lambda_function.main.version
}

output "alias_name" {
  description = "Name of the Lambda alias"
  value       = aws_lambda_alias.main.name
}

output "alias_arn" {
  description = "ARN of the Lambda alias"
  value       = aws_lambda_alias.main.arn
}

output "alias_invoke_arn" {
  description = "Invoke ARN of the Lambda alias"
  value       = aws_lambda_alias.main.invoke_arn
}
