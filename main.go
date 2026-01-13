package main

import (
	"fmt"
	"log"
	"math/rand"
	"roseGelReminder/utils"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bwmarrin/discordgo"
)

const TimestampFile = "timestamp"

func main() {
	env := utils.GetEnv()
	d, err := discordgo.New("Bot " + env.DiscordToken)
	if err != nil {
		panic(err)
	}

	d.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMembers)

	pathPrefix := getPathPrefix()
	log.Printf("running with pathPrefix %s", pathPrefix)

	s, err := utils.CreateS3Datasource(env, pathPrefix)
	if err != nil {
		panic(err)
	}

	err = s.CheckTimeStamp()
	if err != nil {
		panic(fmt.Errorf("CheckTimeStamp error: %v", err))
	}

	config, err := getConfig(s)
	if err != nil {
		panic(err)
	}

	msg, err := fetchMessageContent(s, config)
	if err != nil {
		panic(err)
	}

	data, err, isLeft := constructMessage(env, msg, config, s, d)
	if err != nil {
		panic(err)
	}

	res, err := d.ChannelMessageSendComplex(config.ChannelId, data)
	if err != nil {
		panic(fmt.Sprint("d.ChannelMessageSendComplex error", err))
	}

	log.Printf("Send successful %v", res)
	newTimestamp := res.Timestamp.Format(time.UnixDate)
	err = s.UploadFile(newTimestamp, TimestampFile)
	log.Printf("uploaded timestamp %s", newTimestamp)
	if err != nil {
		panic(fmt.Sprint("timestamp upload error", err))
	}

	if config.IsGel {
		if isLeft {
			err = s.UploadFile("right", config.FileName)
		} else {
			err = s.UploadFile("left", config.FileName)
		}
		if err != nil {
			fmt.Println("Upload errors: ", err)
		}

		log.Printf("uploaded next gel run %s", config.FileName)
	}
}

func getPathPrefix() string {
	now := time.Now()
	return fmt.Sprintf("%02d00", now.Hour())
}

func fetchMessageContent(s *utils.S3DataSource, config *utils.RunConfiguration) (*s3.GetObjectOutput, error) {
	file, err := s.DownloadFile(config.FileName)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile error: %v", err)
	}

	log.Printf("fetchMessageContent downloaded file %s", config.FileName)

	return file, nil
}

func getConfig(s *utils.S3DataSource) (*utils.RunConfiguration, error) {
	config, err := s.DownloadRunConfig()
	if err != nil {
		return nil, fmt.Errorf("DownloadRunConfig error: %v", err)
	}

	if config == nil {
		return nil, fmt.Errorf("no config found")
	}

	log.Println("getConfig downloaded config successfully")

	return config, nil
}

func constructMessage(env *utils.Env, file *s3.GetObjectOutput, config *utils.RunConfiguration, s *utils.S3DataSource, d *discordgo.Session) (data *discordgo.MessageSend, err error, isLeft bool) {
	if config.IsGel {
		content, err := s.ParseResponseToString(file)
		if err != nil {
			return nil, fmt.Errorf("ParseResponseToString error: %v", err), false
		}
		data = &discordgo.MessageSend{Content: fmt.Sprintf("<@%s> %s", config.TagUser, content), AllowedMentions: &discordgo.MessageAllowedMentions{Users: []string{config.TagUser}}}
		isLeft = content == "left"
	} else {
		discordData, err := s.ParseFile(file, config.FileName)
		if err != nil {
			return nil, fmt.Errorf("ParseFile error: %v", err), false
		}
		data = &discordgo.MessageSend{}
		data.Files = make([]*discordgo.File, 0)
		data.Files = append(data.Files, discordData)

		users, err := d.GuildMembers(env.ServerId, "", 1000)
		if err != nil {
			log.Printf("GetGuildMembers error: %v", err)
		}
		log.Printf("GetGuildMembers: found %d users", len(users))
		filteredUserIds := make([]string, 0)
		for _, user := range users {
			canTag := true
			for _, noTag := range config.NoTagList {
				if user.User.ID == noTag {
					canTag = false
				}
			}
			if canTag {
				filteredUserIds = append(filteredUserIds, user.User.ID)
			}
		}

		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		selectedIdx := r.Intn(len(filteredUserIds))

		tagUser := filteredUserIds[selectedIdx]
		log.Printf("Selected user: %s", tagUser)

		data.AllowedMentions = &discordgo.MessageAllowedMentions{Users: []string{tagUser}}
		data.Content = fmt.Sprintf("<@%s>", tagUser)
	}

	log.Println("constructMessage constructed message successfully")

	return data, nil, isLeft
}
