# S3 Bucket Example

This example creates an S3 bucket with:

- Versioning enabled
- Server-side encryption (AES256)
- Public access blocked
- Server access logging configured
- Tags applied

## Usage

This example is used by the InfraSpec S3 tests to verify bucket configuration.

## Resources Created

- `aws_s3_bucket.main` - The main S3 bucket
- `aws_s3_bucket_versioning.main` - Versioning configuration
- `aws_s3_bucket_server_side_encryption_configuration.main` - Encryption configuration
- `aws_s3_bucket_public_access_block.main` - Public access block configuration
- `aws_s3_bucket_logging.main` - Server access logging configuration
- `aws_s3_bucket.logging` - Separate bucket for access logs
- `aws_s3_bucket_public_access_block.logging` - Public access block for logging bucket