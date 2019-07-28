# OCR using GCS and Cloud Functions

Sample code to use google functions with vision API and google cloud storage

## Pipeline

![OCR flow](ocr-flow.jpg)

## Pre setup

- Create pub/sub topic with name `image-file-type`
- Create pub/sub topic with name `pdf-file-type`
- Create pub/sub topic with name `es-upload`

## Deployment
```
gcloud functions deploy DispatchWithFileType \
  --runtime go111 \
  --trigger-resource <target_bucket_name> \
  --trigger-event google.storage.object.finalize
```

```
gcloud functions deploy DetectTextsFromImage \
  --runtime go111 \
  --trigger-topic image-file-types
```
