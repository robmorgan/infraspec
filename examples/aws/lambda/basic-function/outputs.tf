output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.main.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.main.arn
}

output "invoke_arn" {
  description = "Invoke ARN of the Lambda function"
  value       = aws_lambda_function.main.invoke_arn
}

output "role_arn" {
  description = "ARN of the IAM role"
  value       = aws_iam_role.lambda.arn
}

output "runtime" {
  description = "Runtime of the Lambda function"
  value       = aws_lambda_function.main.runtime
}

output "handler" {
  description = "Handler of the Lambda function"
  value       = aws_lambda_function.main.handler
}
