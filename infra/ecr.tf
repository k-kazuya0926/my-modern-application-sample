module "ecr_hello_world" {
  source = "./ecr"

  repository_name = "${local.project_name}-hello-world"
}

module "ecr_tmp" {
  source = "./ecr"

  repository_name = "${local.project_name}-tmp"
}
