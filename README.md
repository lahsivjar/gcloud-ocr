# OCR using GCS and Cloud Functions

Sample code to use google functions with vision API and google cloud storage

## Deploy with
```
gcloud functions deploy DetectTextsFromGCS \
  --runtime go111 \
  --trigger-resource <target_bucket_name> \
  --trigger-event google.storage.object.finalize
```

```
gcloud functions deploy DetectTextsFromImage \
  --runtime go111 \
  --trigger-topic image-file-types
```
