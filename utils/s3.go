package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/bwmarrin/discordgo"
)

const ConfigFile = "config.json"

type S3DataSource struct {
	Client     *s3.Client
	bucket     string
	pathPrefix string
}

type RunConfiguration struct {
	NoTagList []string `json:"no_tag_list"`
	ChannelId string   `json:"channel_id"`
	FileName  string   `json:"file_name"`
	IsGel     bool     `json:"is_gel"`
	TagUser   string   `json:"tag_user"`
}

func CreateS3Datasource(env *Env, pathPrefix string) (*S3DataSource, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	cfg.Region = env.AwsRegion
	client := s3.NewFromConfig(cfg)
	datasource := &S3DataSource{Client: client, bucket: env.S3Bucket, pathPrefix: pathPrefix}
	return datasource, nil
}

func (s S3DataSource) UploadFile(contents string, filename string) error {
	reader := strings.NewReader(contents)
	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.getKey(filename)),
		Body:   reader,
	})
	return err
}

func (s S3DataSource) DownloadFile(filename string) (*s3.GetObjectOutput, error) {
	res, err := s.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.getKey(filename)),
	})
	return res, err
}

func (s S3DataSource) CheckTimeStamp() error {
	key := s.getKey("timestamp")
	res, err := s.DownloadFile(key)
	if err != nil {
		return err
	}

	str, err := s.ParseResponseToString(res)
	if err != nil {
		return err
	}

	timestamp, err := time.Parse(time.UnixDate, str)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	cy, cm, cd := currentTime.Date()
	ty, tm, td := timestamp.Date()

	if cy == ty && cm == tm && cd == td {
		return fmt.Errorf("bot already ran today, bailing")
	}

	return nil
}

// DownloadRunConfig
// config descriptor should be of the format su_m_tu_w_th_f_sa_9
func (s S3DataSource) DownloadRunConfig() (*RunConfiguration, error) {
	res, err := s.DownloadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	str, err := s.ParseResponseToString(res)

	runConfig := &RunConfiguration{}
	err = json.Unmarshal([]byte(str), runConfig)
	if err != nil {
		return nil, err
	}

	return runConfig, nil
}

func (s S3DataSource) GetAllFiles() ([]*discordgo.File, error) {
	files := make([]*discordgo.File, 0)
	keys, err := s.ListAllFilesInFolder()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		file := &discordgo.File{}
		obj, err := s.DownloadFile(key)
		if err != nil {
			log.Println(err)
		}
		file, err = s.ParseFile(obj, key)
		if err != nil {
			log.Println(err)
		}

		files = append(files, file)
	}

	return files, nil

}

func (s S3DataSource) ParseFile(res *s3.GetObjectOutput, key string) (file *discordgo.File, err error) {
	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(res.Body); err != nil {
		return nil, err
	}
	mimetype := *res.ContentType
	file = &discordgo.File{Reader: bytes.NewReader(buffer.Bytes()), Name: key, ContentType: mimetype}

	if err := res.Body.Close(); err != nil {
		return nil, err
	}

	return file, nil
}

func (s S3DataSource) ListAllFilesInFolder() ([]string, error) {
	res, err := s.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.pathPrefix),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %s", err.Error())
	}
	keys := make([]string, 0)

	for _, obj := range res.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

func (s S3DataSource) ParseResponseToString(res *s3.GetObjectOutput) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, res.Body)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s S3DataSource) getKey(fileName string) string {
	return path.Join(s.pathPrefix, fileName)
}

func (s S3DataSource) FileExists(fileName string) bool {
	res, err := s.Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return false
	}

	return res.ContentType == nil
}
