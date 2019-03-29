package main

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yansal/img/img"
	"github.com/yansal/img/storage/backends/s3"
)

type apiGatewayRequest struct {
	QueryStringParameters map[string]string
}

type apiGatewayResponse struct {
	Body            string            `json:"body"`
	Headers         map[string]string `json:"headers"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

func newPayload(params map[string]string) (img.Payload, error) {
	var payload img.Payload
	path := params["path"]
	if path == "" {
		return payload, errors.New("path is required")
	}
	payload.Path = path

	width := params["width"]
	if width != "" {
		w, err := strconv.Atoi(width)
		if err != nil {
			return payload, err
		}
		payload.Width = w
	}

	height := params["height"]
	if height != "" {
		w, err := strconv.Atoi(height)
		if err != nil {
			return payload, err
		}
		payload.Height = w
	}
	return payload, nil
}

func HandleRequest(ctx context.Context, req apiGatewayRequest) (*apiGatewayResponse, error) {
	bucket := os.Getenv("S3BUCKET")
	if bucket == "" {
		return nil, errors.New("S3BUCKET env is required")
	}

	storage, err := s3.New(bucket)
	if err != nil {
		return nil, err
	}

	payload, err := newPayload(req.QueryStringParameters)
	b, err := img.NewProcessor(storage).Process(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &apiGatewayResponse{
		Body: base64.StdEncoding.EncodeToString(b),
		Headers: map[string]string{
			"Content-Type": http.DetectContentType(b),
		},
		IsBase64Encoded: true,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
