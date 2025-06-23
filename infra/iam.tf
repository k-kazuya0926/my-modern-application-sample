data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

module "hello_world_role" {
  source     = "./iam_role"
  base_name  = "${local.project_name}-hello-world"
  policy     = data.aws_iam_policy_document.hello_world.json
  identifier = "lambda.amazonaws.com"
}

data "aws_iam_policy_document" "hello_world" {
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
    resources = ["arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${local.project_name}-hello-world:*"]
  }
}


module "tmp_role" {
  source     = "./iam_role"
  base_name  = "${local.project_name}-tmp"
  policy     = data.aws_iam_policy_document.tmp.json
  identifier = "lambda.amazonaws.com"
}

data "aws_iam_policy_document" "tmp" {
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
    resources = ["arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${local.project_name}-tmp:*"]
  }
}


resource "aws_iam_openid_connect_provider" "github_actions" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = [
    "sts.amazonaws.com"
  ]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"
  ]

  tags = {
    Name = "${local.project_name}-github-actions-oidc"
  }
}

resource "aws_iam_role" "github_actions" {
  name               = "${local.project_name}-github-actions-role"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role.json
}

data "aws_iam_policy_document" "github_actions_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github_actions.arn]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${local.github_owner_name}/${local.project_name}:*"]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "github_actions" {
  name   = "${local.project_name}-github-actions-policy"
  policy = data.aws_iam_policy_document.github_actions.json
}

data "aws_iam_policy_document" "github_actions" {
  statement {
    sid    = "ECRPermissions"
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "ECRRepositoryPermissions"
    effect = "Allow"
    actions = [
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:CompleteLayerUpload",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:PutImage",
      "ecr:UploadLayerPart"
    ]
    resources = [
      "arn:aws:ecr:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:repository/${local.project_name}-hello-world",
      "arn:aws:ecr:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:repository/${local.project_name}-tmp"
    ]
  }

  statement {
    sid    = "LambdaDeployPermissions"
    effect = "Allow"
    actions = [
      "lambda:UpdateFunctionCode",
      "lambda:GetFunction",
      "lambda:GetFunctionConfiguration",
      "lambda:PublishVersion",
      "lambda:UpdateAlias"
    ]
    resources = [
      "arn:aws:lambda:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:function:${local.project_name}-hello-world",
      "arn:aws:lambda:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:function:${local.project_name}-hello-world:*",
      "arn:aws:lambda:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:function:${local.project_name}-tmp",
      "arn:aws:lambda:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:function:${local.project_name}-tmp:*"
    ]
  }
}

resource "aws_iam_role_policy_attachment" "github_actions" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.github_actions.arn
}
