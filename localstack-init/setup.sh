awslocal s3 mb s3://factory-access-log-bucket --region eu-west-1
awslocal s3api put-bucket-acl --bucket factory-access-log-bucket --acl public-read
awslocal iam create-role --role-name super-role --assume-role-policy-document file:///docker-entrypoint-initaws.d/iam_policy.json
awslocal firehose create-delivery-stream --region eu-west-1 --cli-input-json file:///docker-entrypoint-initaws.d/firehose_config.json
