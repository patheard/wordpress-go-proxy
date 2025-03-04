# WordPress Go Proxy
Proxy function that retrieves and renders WordPress content using the [GC Design System](https://design-system.alpha.canada.ca/).

## Setup
This sets up a simple AWS Lambda function with a function URL.

```sh
# Build the Lambda function's docker image.  The terraform expects it to
# be ARM64, but this can be updated by adjusting the function's `architectures`.
docker build --tag wordpress-go-proxy --platform linux/arm64 .

cd terraform
terraform init
terraform apply
```

:warning: The first Terraform apply will fail since the Docker image won't be in the new ECR yet.  Push up the Docker image and re-run `terraform apply` to fix.