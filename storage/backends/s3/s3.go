package s3

import (
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage struct {
	s3     *s3.S3
	bucket string
}

func New(bucket string) (*Storage, error) {
	s, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return &Storage{
		bucket: bucket,
		s3:     s3.New(s),
	}, nil
}

func (s *Storage) Get(path string) ([]byte, error) {
	out, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return ioutil.ReadAll(out.Body)
}
