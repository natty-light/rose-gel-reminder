GOOS=linux GOARCH=amd64 go build -o main main.go
zip main.zip main
aws lambda update-function-code --function-name rose-gel-reminder --zip-file fileb://main.zip --profile default --region us-east-1