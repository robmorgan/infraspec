# DynamoDB Table Autoscaling example

This example creates a AWS DynamoDB table with Autoscaling enabled. Be sure to read [the note]([../../README.md#Notes](https://github.com/terraform-aws-modules/terraform-aws-dynamodb-table/blob/master/README.md))
about autoscaling settings causing the table to be recreated.

## Usage

To run this example you need to execute:

```bash
$ terraform init
$ terraform plan
$ terraform apply
```

Note that this example may create resources which can cost money (AWS Elastic IP, for example). Run `terraform destroy` when you don't need these resources.
