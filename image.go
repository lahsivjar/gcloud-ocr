package ocr

import (
	"context"
	"fmt"
	"log"

	"github.com/lahsivjar/gcloud-ocr/shared/models"
	"github.com/lahsivjar/gcloud-ocr/shared/ocrpubsub"

	vision "cloud.google.com/go/vision/apiv1"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

// DetectTextsFromImage will use gcloud vision API to detect texts from image
func DetectTextsFromImage(ctx context.Context, m models.PubSubMessage) error {
	g := models.GCSEvent{}
	err := g.Decode(m.Data)
	if err != nil {
		return err
	}

	log.Printf("Bucket: %v\n", g.Bucket)
	log.Printf("File: %v\n", g.Name)

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return fmt.Errorf("vision.NewImageAnnotatorClient: %v", err)
	}

	imgSource := &visionpb.Image{
		Source: &visionpb.ImageSource{
			ImageUri: fmt.Sprintf("gs://%s/%s", g.Bucket, g.Name),
		},
	}

	annotations, err := client.DetectTexts(ctx, imgSource, nil, 10)
	if err != nil {
		return fmt.Errorf("failed to detect document text: %v", err)
	}

	var detectedTexts []string
	if len(annotations) > 0 {
		for i := range annotations {
			a := annotations[i]
			detectedTexts = append(detectedTexts, a.Description)
		}
	}

	log.Printf("Detected texts: %v\n", detectedTexts)

	ep := models.ESPayload{
		Event:         g,
		DetectedTexts: detectedTexts,
	}

	msg, err := ep.Encode()
	if err != nil {
		return err
	}

	sID, err := ocrpubsub.Publish(ctx, ocrpubsub.ESUploadTopic, msg)
	if err != nil {
		return fmt.Errorf("failed to publish message with error %v", err)
	}

	log.Printf("Published a message with msg ID: %v\n", sID)
	return nil
}
