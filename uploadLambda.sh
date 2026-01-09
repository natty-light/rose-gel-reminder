GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip bootstrap.zip bootstrap
aws lambda update-function-code --function-name rose-gel-reminder --zip-file fileb://bootstrap.zip --profile default --region us-east-2