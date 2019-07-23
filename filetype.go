package ocr

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"strings"

	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/lahsivjar/gcloud-ocr/shared/models"
)

const (
	imageFileTopic string = "image-file-types"
	pdfFileTopic   string = "pdf-file-types"
)

// DispatchWithFileType gets the file type metadata info from GCS and dispatches
// to correct pub/sub topic
func DispatchWithFileType(ctx context.Context, e models.GCSEvent) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}

	log.Printf("Event ID: %v\n", meta.EventID)
	log.Printf("Event type: %v\n", meta.EventType)
	log.Printf("Bucket: %v\n", e.Bucket)
	log.Printf("File: %v\n", e.Name)

	client, err := storage.NewClient(ctx)
	bucket := client.Bucket(e.Bucket)
	rc, err := bucket.Object(e.Name).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("bucket.Object.NewReader: %v", err)
	}

	cType := rc.ContentType()
	log.Printf("Content type: %v\n", cType)

	mime, _, err := mime.ParseMediaType(cType)
	if err != nil {
		return fmt.Errorf("failed to find mime type of file, aborting with %v", err)
	}

	var topic string
	if strings.HasPrefix(mime, "image") {
		topic = imageFileTopic
	} else if mime == "application/pdf" {
		topic = pdfFileTopic
	} else {
		return fmt.Errorf("mimetype %s is not supported", mime)
	}
	sID, err := publish(ctx, imageFileTopic, e)
	if err != nil {
		return fmt.Errorf("failed to publish message on topic %s with error %v", topic, err)
	}

	log.Printf("Published a message on topic %s with msg ID: %v\n", topic, sID)

	return nil
}

func publish(ctx context.Context, topic string, e models.GCSEvent) (string, error) {
	gcpProject := os.Getenv("GCP_PROJECT")
	if gcpProject == "" {
		return "", fmt.Errorf("failed to retrieve gcp project id")
	}

	log.Printf("Project ID: %v\n", gcpProject)
	client, err := pubsub.NewClient(ctx, gcpProject)
	if err != nil {
		return "", fmt.Errorf("pubsub.NewClient: %v", err)
	}

	t := client.Topic(topic)
	msg, err := e.Encode()
	if err != nil {
		return "", fmt.Errorf("pubsub.NewClient: %v", err)
	}

	r := t.Publish(ctx, &pubsub.Message{
		Data: msg.Bytes(),
	})

	id, err := r.Get(ctx)
	if err != nil {
		return "", err
	}

	return id, nil
}
