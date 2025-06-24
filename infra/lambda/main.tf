resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = var.execution_role_arn
  package_type  = "Image"
  image_uri     = var.image_uri

  # メモリとタイムアウトの設定
  memory_size = var.memory_size
  timeout     = var.timeout

  # 環境変数
  dynamic "environment" {
    for_each = length(var.environment_variables) > 0 ? [1] : []
    content {
      variables = var.environment_variables
    }
  }

  # VPC設定（オプション）
  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }

  # デッドレターキュー設定（オプション）
  dynamic "dead_letter_config" {
    for_each = var.dead_letter_queue_arn != null ? [1] : []
    content {
      target_arn = var.dead_letter_queue_arn
    }
  }

  # 予約済み並行実行数（オプション）
  reserved_concurrent_executions = var.reserved_concurrent_executions

  # ログ保持期間
  depends_on = [
    aws_cloudwatch_log_group.lambda_logs,
  ]

  tags = var.tags
}

# CloudWatch Logs グループ
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# Lambda関数のエイリアス（オプション）
resource "aws_lambda_alias" "this" {
  count = var.create_alias ? 1 : 0

  name             = var.alias_name
  description      = var.alias_description
  function_name    = aws_lambda_function.this.function_name
  function_version = aws_lambda_function.this.version
}

# Lambda関数のバージョン（オプション）
resource "aws_lambda_function_event_invoke_config" "this" {
  count = var.maximum_retry_attempts != null || var.maximum_event_age != null ? 1 : 0

  function_name = aws_lambda_function.this.function_name

  maximum_retry_attempts       = var.maximum_retry_attempts
  maximum_event_age_in_seconds = var.maximum_event_age

  dynamic "destination_config" {
    for_each = var.destination_config != null ? [var.destination_config] : []
    content {
      dynamic "on_failure" {
        for_each = destination_config.value.on_failure != null ? [destination_config.value.on_failure] : []
        content {
          destination = on_failure.value.destination
        }
      }
      dynamic "on_success" {
        for_each = destination_config.value.on_success != null ? [destination_config.value.on_success] : []
        content {
          destination = on_success.value.destination
        }
      }
    }
  }
}
