module "lambda_hello_world" {
  source = "./lambda"

  function_name      = "${local.project_name}-hello-world"
  execution_role_arn = module.hello_world_role.iam_role_arn
  image_uri          = "${module.ecr_hello_world.repository_url}:dummy"
}

module "lambda_tmp" {
  source = "./lambda"

  function_name      = "${local.project_name}-tmp"
  execution_role_arn = module.tmp_role.iam_role_arn
  image_uri          = "${module.ecr_tmp.repository_url}:dummy"
}
