output "function_name" {
  description = "Lambda関数の名前"
  value       = aws_lambda_function.this.function_name
}

output "function_arn" {
  description = "Lambda関数のARN"
  value       = aws_lambda_function.this.arn
}

output "function_qualified_arn" {
  description = "Lambda関数の修飾ARN（バージョン付き）"
  value       = aws_lambda_function.this.qualified_arn
}

output "function_invoke_arn" {
  description = "Lambda関数の呼び出しARN"
  value       = aws_lambda_function.this.invoke_arn
}

output "function_version" {
  description = "Lambda関数のバージョン"
  value       = aws_lambda_function.this.version
}

output "function_last_modified" {
  description = "Lambda関数の最終更新日時"
  value       = aws_lambda_function.this.last_modified
}

output "function_source_code_hash" {
  description = "Lambda関数のソースコードハッシュ"
  value       = aws_lambda_function.this.source_code_hash
}

output "function_source_code_size" {
  description = "Lambda関数のソースコードサイズ"
  value       = aws_lambda_function.this.source_code_size
}

output "cloudwatch_log_group_name" {
  description = "CloudWatch Logsグループの名前"
  value       = aws_cloudwatch_log_group.lambda_logs.name
}

output "cloudwatch_log_group_arn" {
  description = "CloudWatch LogsグループのARN"
  value       = aws_cloudwatch_log_group.lambda_logs.arn
}

output "alias_arn" {
  description = "Lambda関数エイリアスのARN（作成された場合）"
  value       = var.create_alias ? aws_lambda_alias.this[0].arn : null
}

output "alias_name" {
  description = "Lambda関数エイリアスの名前（作成された場合）"
  value       = var.create_alias ? aws_lambda_alias.this[0].name : null
}

output "alias_invoke_arn" {
  description = "Lambda関数エイリアスの呼び出しARN（作成された場合）"
  value       = var.create_alias ? aws_lambda_alias.this[0].invoke_arn : null
}
