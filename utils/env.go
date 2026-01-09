package utils

import (
	"os"
)

type Env struct {
	DiscordToken string
	AwsRegion    string
	S3Bucket     string
	ServerId     string
}

func GetEnv() *Env {
	env := &Env{}

	env.DiscordToken = os.Getenv("DISCORD_TOKEN")
	env.AwsRegion = os.Getenv("BUCKET_REGION")
	env.S3Bucket = os.Getenv("AWS_S3_BUCKET")
	env.ServerId = os.Getenv("DISCORD_SERVER_ID")

	return env
}
