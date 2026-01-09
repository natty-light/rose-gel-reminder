package main

import (
	"fmt"
	"path"
	"roseGelReminder/utils"
	"strings"
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
	pathPrefix := getPathPrefix()

	s, err := utils.CreateS3Datasource(&env, pathPrefix)
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

	msg, err := fetchMessageContent(s, *config)
	if err != nil {
		panic(err)
	}

	data, err, isLeft := constructMessage(env, msg, *config, s)
	if err != nil {
		panic(err)
	}

	res, err := d.ChannelMessageSendComplex(env.ChannelId, data)
	if err != nil {
		panic(fmt.Sprint("d.ChannelMessageSendComplex error", err))
	}

	fmt.Println("Send successful", res)
	err = s.UploadFile(res.Timestamp.Format(time.UnixDate), TimestampFile)
	if err != nil {
		panic(fmt.Sprint("timestamp upload error", err))
	}

	if config.IsGel {
		key := path.Join(pathPrefix, config.FileName)
		if isLeft {
			err = s.UploadFile("right", key)
		} else {
			err = s.UploadFile("left", key)
		}
		if err != nil {
			fmt.Println("Upload errors: ", err)
		}
	}
}

func getPathPrefix() string {
	now := time.Now()
	weekday := strings.ToLower(now.Weekday().String())
	hour := fmt.Sprintf("%02d00", now.Hour())
	return path.Join(weekday, hour)
}

func fetchMessageContent(s *utils.S3DataSource, config utils.RunConfiguration) (*s3.GetObjectOutput, error) {
	file, err := s.DownloadFile(config.FileName)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile error: %v", err)
	}

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

	return config, nil
}

func constructMessage(env utils.Env, file *s3.GetObjectOutput, config utils.RunConfiguration, s *utils.S3DataSource) (data *discordgo.MessageSend, err error, isLeft bool) {
	if config.IsGel {
		content, err := s.ParseResponseToString(file)
		if err != nil {
			return nil, fmt.Errorf("ParseResponseToString error: %v", err), false
		}
		data = &discordgo.MessageSend{Content: fmt.Sprintf("<@%s> %s", env.RoseUserId, content), AllowedMentions: &discordgo.MessageAllowedMentions{Users: []string{env.RoseUserId}}}
		isLeft = content == "left"
	} else {
		discordData, err := s.ParseFile(file, config.FileName)
		if err != nil {
			return nil, fmt.Errorf("ParseFile error: %v", err), false
		}
		data = &discordgo.MessageSend{}
		data.Files = make([]*discordgo.File, 0)
		data.Files = append(data.Files, discordData)
	}

	return data, nil, isLeft
}
