package utils

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

type S3DataSource struct {
	Client *s3.Client
}

func CreateS3Client(env Env) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg.Region = env.AwsRegion
	if err != nil {
		return nil, err
	}
	Client := s3.NewFromConfig(cfg)
	return Client, nil
}

func (s S3DataSource) UploadFile(env Env, contents string, filename string) error {
	reader := strings.NewReader(contents)
	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(env.S3Bucket),
		Key:    aws.String(filename),
		Body:   reader})
	return err
}

func (s S3DataSource) DownloadFile(env Env, filename string) (string, error) {
	res, err := s.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(env.S3Bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return "", err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return "", nil
	}

	return buf.String(), nil
}

func (s S3DataSource) CheckTimeStamp(env Env) error {
	str, err := s.DownloadFile(env, "timestamp")
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
