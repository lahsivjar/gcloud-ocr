package ocrpubsub

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
)

func Publish(ctx context.Context, topic string, msg bytes.Buffer) (string, error) {
	gcpProject, err := GetGcpProjectID()
	if err != nil {
		return "", err
	}

	log.Printf("Project ID: %v\n", gcpProject)

	client, err := pubsub.NewClient(ctx, gcpProject)
	if err != nil {
		return "", fmt.Errorf("pubsub.NewClient: %v", err)
	}

	t := client.Topic(topic)
	r := t.Publish(ctx, &pubsub.Message{
		Data: msg.Bytes(),
	})
	id, err := r.Get(ctx)
	if err != nil {
		return "", err
	}

	return id, nil
}

func GetGcpProjectID() (string, error) {
	gcpProject := os.Getenv("GCP_PROJECT")
	if gcpProject == "" {
		return "", fmt.Errorf("failed to retrieve gcp project id")
	}

	return gcpProject, nil
}
