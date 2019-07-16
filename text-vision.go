package ocr

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/functions/metadata"
	"github.com/lahsivjar/gcloud-ocr/pkg/models"

	vision "cloud.google.com/go/vision/apiv1"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

// DetectTextsFromGCS will use gcloud vision API to detect texts from image
func DetectTextsFromGCS(ctx context.Context, e models.GCSEvent) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}

	log.Printf("Event ID: %v\n", meta.EventID)
	log.Printf("Event type: %v\n", meta.EventType)
	log.Printf("Bucket: %v\n", e.Bucket)
	log.Printf("File: %v\n", e.Name)

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return fmt.Errorf("vision.NewImageAnnotatorClient: %v", err)
	}

	imgSource := &visionpb.Image{
		Source: &visionpb.ImageSource{
			ImageUri: fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name),
		},
	}

	text, err := client.DetectTexts(ctx, imgSource, nil, 10)
	if err != nil {
		return fmt.Errorf("failed to detect document text: %v", err)
	}

	log.Printf("detected text: %v\n", text)

	return nil
}
