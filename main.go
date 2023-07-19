package main

import (
	"fmt"
	"roseGelReminder/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

const TIMESTAMP_FILE = "timestamp"
const NEXT_DAY_MESSAGE = "next"

func main() {
	env := utils.GetEnv()
	s, err := discordgo.New("Bot " + env.DiscordToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := utils.CreateS3Client(env)
	if err != nil {
		fmt.Println("S3 Client config error", err)
		return
	}

	d := utils.S3DataSource{Client: client}
	err = d.CheckTimeStamp(env)
	if err != nil {
		fmt.Println(err)
		return
	}

	today, err := d.DownloadFile(env, NEXT_DAY_MESSAGE)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := discordgo.MessageSend{Content: fmt.Sprintf("<@%s> %s", env.RoseUserId, today), AllowedMentions: &discordgo.MessageAllowedMentions{Users: []string{env.RoseUserId}}}

	res, err := s.ChannelMessageSendComplex(env.ChannelId, &data)
	if err != nil {
		fmt.Println("s.ChannelMessageSendComplex error", err)
	} else {
		fmt.Println("Send successful", res)
		err = d.UploadFile(env, res.Timestamp.Format(time.UnixDate), TIMESTAMP_FILE)
		if err != nil {
			fmt.Println("Upload error: ", err)
		} else {
			// Upload opposite of what we sent today
			if today == "left" {
				err = d.UploadFile(env, "right", NEXT_DAY_MESSAGE)
			} else if today == "right" {
				err = d.UploadFile(env, "left", NEXT_DAY_MESSAGE)
			}
			if err != nil {
				fmt.Println("Upload errors: ", err)
			}
		}
	}
}
