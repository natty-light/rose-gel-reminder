package utils

import (
	"os"
)

type Env struct {
	DiscordToken string
	ChannelId    string
	AwsRegion    string
	S3Bucket     string
	RoseUserId   string
}

func GetEnv() Env {
	var env Env = Env{}

	env.DiscordToken = os.Getenv("DISCORD_TOKEN")
	env.ChannelId = os.Getenv("DISCORD_CHANNEL_ID")
	env.AwsRegion = os.Getenv("BUCKET_REGION")
	env.S3Bucket = os.Getenv("AWS_S3_BUCKET")
	env.RoseUserId = os.Getenv("ROSE_USER_ID")

	return env
}
