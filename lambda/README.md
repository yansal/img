# Build and upload

```
GOOS=linux go build -o img && zip img.zip img && aws s3 cp img.zip s3://<my-lambda-bucket>
```

# Update lambda function code

```
aws lambda update-function-code --function-name img --s3-bucket <my-lambda-bucket> --s3-key img.zip
```

# Required permission for lambda role

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::<my-img-bucket>/*"
        }
    ]
}
```