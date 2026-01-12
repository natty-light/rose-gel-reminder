GOOS=linux GOARCH=amd64 go build -o main main.go
zip main.zip main
aws lambda create-function --function-name rose-gel-reminder-alarm --runtime p --region us-east-1