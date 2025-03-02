variable "billing_code" {
  description = "The billing code to use for cost allocation"
  type        = string
}

variable "env" {
  description = "The target environment"
  type        = string
}

variable "wordpress_menu_id_en" {
  description = "The English menu to display"
  type        = string
}

variable "wordpress_menu_id_fr" {
  description = "The French menu to display"
  type        = string
}

variable "region" {
  description = "The AWS region to deploy resources"
  type        = string
}

variable "site_name_en" {
  description = "The English site name"
  type        = string
}

variable "site_name_fr" {
  description = "The French site name"
  type        = string
}

variable "wordpress_url" {
  description = "The URL of the WordPress site"
  type        = string
}

variable "wordpress_username" {
  description = "The WordPress user"
  type        = string
  sensitive   = true
}

variable "wordpress_password" {
  description = "The WordPress user's app password"
  type        = string
  sensitive   = true
}