resource "aws_ecr_repository" "wordpress_go_proxy" {
  name                 = "wordpress-go-proxy"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.common_tags
}