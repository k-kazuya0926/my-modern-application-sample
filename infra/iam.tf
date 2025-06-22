data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

module "my-modern-application-sample-hello-world-role" {
  source     = "./iam_role"
  name       = "my-modern-application-sample-hello-world-role"
  policy     = data.aws_iam_policy_document.my-modern-application-sample-hello-world-role.json
  identifier = "lambda.amazonaws.com"
}

data "aws_iam_policy_document" "my-modern-application-sample-hello-world-role" {
  statement {
    effect    = "Allow"
    actions   = ["logs:CreateLogGroup"]
    resources = ["arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/my-modern-application-sample-hello-world:*"]
  }
}

module "my-modern-application-sample-tmp-role" {
  source     = "./iam_role"
  name       = "my-modern-application-sample-tmp-role"
  policy     = data.aws_iam_policy_document.my-modern-application-sample-tmp-role.json
  identifier = "lambda.amazonaws.com"
}

data "aws_iam_policy_document" "my-modern-application-sample-tmp-role" {
  statement {
    effect    = "Allow"
    actions   = ["logs:CreateLogGroup"]
    resources = ["arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/my-modern-application-sample-tmp:*"]
  }
}
