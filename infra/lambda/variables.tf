variable "function_name" {
  description = "Lambda関数の名前"
  type        = string
}

variable "execution_role_arn" {
  description = "Lambda実行ロールのARN"
  type        = string
}

variable "image_uri" {
  description = "コンテナイメージのURI（ECRリポジトリのURI）"
  type        = string
}

variable "memory_size" {
  description = "Lambda関数のメモリサイズ（MB）"
  type        = number
  default     = 128
}

variable "timeout" {
  description = "Lambda関数のタイムアウト（秒）"
  type        = number
  default     = 3
}

variable "environment_variables" {
  description = "環境変数のマップ"
  type        = map(string)
  default     = {}
}

variable "vpc_config" {
  description = "VPC設定"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default = null
}

variable "dead_letter_queue_arn" {
  description = "デッドレターキューのARN"
  type        = string
  default     = null
}

variable "reserved_concurrent_executions" {
  description = "予約済み並行実行数"
  type        = number
  default     = null
}

variable "log_retention_days" {
  description = "CloudWatch Logsの保持期間（日）"
  type        = number
  default     = 7

  validation {
    condition = contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653
    ], var.log_retention_days)
    error_message = "log_retention_days は有効なCloudWatch Logsの保持期間である必要があります。"
  }
}

variable "create_alias" {
  description = "Lambda関数のエイリアスを作成するかどうか"
  type        = bool
  default     = false
}

variable "alias_name" {
  description = "Lambda関数のエイリアス名"
  type        = string
  default     = "LIVE"
}

variable "alias_description" {
  description = "Lambda関数のエイリアスの説明"
  type        = string
  default     = "Lambda function alias"
}

variable "maximum_retry_attempts" {
  description = "非同期呼び出しの最大再試行回数"
  type        = number
  default     = null

  validation {
    condition     = var.maximum_retry_attempts == null || (var.maximum_retry_attempts >= 0 && var.maximum_retry_attempts <= 2)
    error_message = "maximum_retry_attempts は 0 から 2 の間である必要があります。"
  }
}

variable "maximum_event_age" {
  description = "イベントの最大保持時間（秒）"
  type        = number
  default     = null

  validation {
    condition     = var.maximum_event_age == null || (var.maximum_event_age >= 60 && var.maximum_event_age <= 21600)
    error_message = "maximum_event_age は 60 から 21600 秒の間である必要があります。"
  }
}

variable "destination_config" {
  description = "非同期呼び出しの送信先設定"
  type = object({
    on_failure = optional(object({
      destination = string
    }))
    on_success = optional(object({
      destination = string
    }))
  })
  default = null
}

variable "tags" {
  description = "リソースに適用するタグ"
  type        = map(string)
  default     = {}
}
