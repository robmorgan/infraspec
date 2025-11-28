output "role_name" {
  description = "Name of the IAM role"
  value       = aws_iam_role.main.name
}

output "role_arn" {
  description = "ARN of the IAM role"
  value       = aws_iam_role.main.arn
}

output "role_id" {
  description = "ID of the IAM role"
  value       = aws_iam_role.main.id
}

output "policy_arn" {
  description = "ARN of the IAM policy"
  value       = aws_iam_policy.main.arn
}

output "policy_name" {
  description = "Name of the IAM policy"
  value       = aws_iam_policy.main.name
}

output "instance_profile_name" {
  description = "Name of the IAM instance profile"
  value       = aws_iam_instance_profile.main.name
}

output "instance_profile_arn" {
  description = "ARN of the IAM instance profile"
  value       = aws_iam_instance_profile.main.arn
}
