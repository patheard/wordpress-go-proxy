module "wordpress_go_proxy_lambda" {
  source    = "github.com/cds-snc/terraform-modules//lambda?ref=v10.3.0"
  name      = "wordpress-go-proxy"
  ecr_arn   = aws_ecr_repository.wordpress_go_proxy.arn
  image_uri = "${aws_ecr_repository.wordpress_go_proxy.repository_url}:latest"

  architectures          = ["arm64"]
  memory                 = 1024
  timeout                = 10
  enable_lambda_insights = true

  environment_variables = {
    SITE_NAME_EN         = var.site_name_en
    SITE_NAME_FR         = var.site_name_fr
    WORDPRESS_MENU_ID_EN = var.wordpress_menu_id_en
    WORDPRESS_MENU_ID_FR = var.wordpress_menu_id_fr
    WORDPRESS_URL        = var.wordpress_url
    WORDPRESS_USERNAME   = var.wordpress_username
    WORDPRESS_PASSWORD   = var.wordpress_password
  }

  billing_tag_value = var.billing_code
}

resource "aws_lambda_function_url" "wordpress_go_proxy_lambda" {
  function_name      = module.wordpress_go_proxy_lambda.function_name
  authorization_type = "NONE"
}
