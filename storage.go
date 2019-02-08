package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-redis/redis"
)

type storage interface {
	Get(path string) ([]byte, error)
	Set(path string, data []byte) error
}

type local struct{ Base string }

func (l *local) Get(path string) ([]byte, error) {
	abs := filepath.Join(l.Base, path)
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, nil
	}
	return ioutil.ReadFile(abs)
}

func (l *local) Set(path string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(l.Base, path), data, 0644)
}

type redisS3 struct {
	redis  *redis.Client
	s3     *s3.S3
	bucket string
}

func (r *redisS3) Get(path string) ([]byte, error) {
	_, err := r.redis.Get(path).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	out, err := r.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return ioutil.ReadAll(out.Body)
}

func (r *redisS3) Set(path string, data []byte) error {
	_, err := r.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return err
	}
	return r.redis.Set(path, "", 0).Err()
}

func newredis() *redis.Client {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://:6379"
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	poolsize, _ := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
	opts.PoolSize = poolsize
	client := redis.NewClient(opts)
	if err := client.Ping().Err(); err != nil {
		panic(err)
	}
	return client
}

func news3() *s3.S3 {
	return s3.New(session.Must(session.NewSession()))
}
