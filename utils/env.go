package utils

import (
	"os"
)

type Env struct {
	DiscordToken string
	AwsRegion    string
	S3Bucket     string
}

func GetEnv() Env {
	var env Env = Env{}

	env.DiscordToken = os.Getenv("DISCORD_TOKEN")
	env.AwsRegion = os.Getenv("BUCKET_REGION")
	env.S3Bucket = os.Getenv("AWS_S3_BUCKET")

	return env
}
