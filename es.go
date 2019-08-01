package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/lahsivjar/gcloud-ocr/shared/iap"
	"github.com/lahsivjar/gcloud-ocr/shared/models"
	"golang.org/x/oauth2"
)

const (
	esURLKey      = "ES_REQUEST_URL"
	clientIDKey   = "OAUTH_CLIENT_ID"
	svcAccountKey = "SERVICE_ACCOUNT_CREDS"
)

// UploadTextsToES will upload to ES (running on k8s and exposed via IAP)
func UploadTextsToES(ctx context.Context, m models.PubSubMessage) error {
	ep := models.ESPayload{}
	err := ep.Decode(m.Data)
	if err != nil {
		return err
	}

	log.Printf("Bucket: %v\n", ep.Event.Bucket)
	log.Printf("File: %v\n", ep.Event.Name)

	req, err := getHTTPRequest(ctx, ep)
	if err != nil {
		return err
	}

	httpClient := &http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	log.Printf("Response status: %v\n", res.Status)

	resBody, _ := ioutil.ReadAll(res.Body)
	log.Printf("Response body: %v\n", string(resBody))

	return nil
}

func getHTTPRequest(ctx context.Context, ep models.ESPayload) (*http.Request, error) {
	esURL, err := getEsURL()
	if err != nil {
		return nil, err
	}

	token, err := getBearerToken(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully generated bearer token")

	tenantID, err := extractTenantID(ep.Event.Name)
	if err != nil {
		return nil, err
	}

	reqURL := esURL + "/" + ep.Event.Name
	reqBody, err := json.Marshal(ep.DetectedTexts)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant", tenantID)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func getBearerToken(ctx context.Context) (string, error) {
	oauthID, err := getOauthClientID()
	if err != nil {
		return "", err
	}

	svcAccount, err := getServiceAccount()
	if err != nil {
		return "", err
	}

	hc := oauth2.NewClient(ctx, nil)
	iap, err := iap.New(hc, svcAccount, oauthID)
	if err != nil {
		return "", err
	}

	token, err := iap.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get iap token with error: %v", err)
	}

	return token, nil
}

func getEsURL() (string, error) {
	es := os.Getenv(esURLKey)
	if es == "" {
		return "", errors.New("failed to get elasticsearch url")
	}

	return es, nil
}

func getOauthClientID() (string, error) {
	oauthID := os.Getenv(clientIDKey)
	if oauthID == "" {
		return "", errors.New("failed to get oauth client id")
	}

	return oauthID, nil
}

func getServiceAccount() (string, error) {
	svcAcc := os.Getenv(svcAccountKey)
	if svcAcc == "" {
		return "", errors.New("failed to get service account creds")
	}

	data, err := base64.StdEncoding.DecodeString(svcAcc)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func extractTenantID(fileName string) (string, error) {
	fileInfo := strings.Split(fileName, "/")
	if len(fileInfo) != 2 {
		return "", fmt.Errorf("invalid filename found: %v", fileName)
	}

	return fileInfo[0], nil
}
