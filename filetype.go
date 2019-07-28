package ocr

import (
	"context"
	"fmt"
	"log"
	"mime"
	"strings"

	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
	"github.com/lahsivjar/gcloud-ocr/shared/models"
	"github.com/lahsivjar/gcloud-ocr/shared/ocrpubsub"
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
		topic = ocrpubsub.ImageFileTopic
	} else if mime == "application/pdf" {
		topic = ocrpubsub.PdfFileTopic
	} else {
		return fmt.Errorf("mimetype %s is not supported", mime)
	}
	msg, err := e.Encode()
	if err != nil {
		return err
	}

	sID, err := ocrpubsub.Publish(ctx, topic, msg)
	if err != nil {
		return fmt.Errorf("failed to publish message on topic %s with error %v", topic, err)
	}

	log.Printf("Published a message on topic %s with msg ID: %v\n", topic, sID)

	return nil
}
