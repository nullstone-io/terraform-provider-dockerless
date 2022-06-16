data "aws_caller_identity" "this" {}

data "aws_ecr_authorization_token" "temporary" {
  registry_id = data.aws_caller_identity.this.account_id
}

provider "dockerless" {
  registry_auth {
    address  = data.aws_ecr_authorization_token.temporary.proxy_endpoint
    username = data.aws_ecr_authorization_token.temporary.user_name
    password = data.aws_ecr_authorization_token.temporary.password
  }
}
