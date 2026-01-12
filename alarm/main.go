package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Env struct {
	DiscordToken     string
	DeveloperUserId  string
	SupportChannelId string
}

func GetEnv() *Env {
	env := &Env{}

	env.DiscordToken = os.Getenv("DISCORD_TOKEN")
	env.DeveloperUserId = os.Getenv("DISCORD_DEVELOPER_ID")
	env.SupportChannelId = os.Getenv("DISCORD_SUPPORT_CHANNEL_ID")

	return env
}

func main() {
	env := GetEnv()
	d, err := discordgo.New("Bot " + env.DiscordToken)
	if err != nil {
		log.Fatal(err)
		return
	}

	data := &discordgo.MessageSend{Content: fmt.Sprintf("<@%s> rose-gel-reminder lambda in alarm %v", env.DeveloperUserId, time.Now().Format("2006-01-15:15:04:05")), AllowedMentions: &discordgo.MessageAllowedMentions{Users: []string{env.DeveloperUserId}}}
	_, err = d.ChannelMessageSendComplex(env.SupportChannelId, data)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("notified support developer")
}
