try:
    bucket = storage_client.get_bucket(bucket_name)
    bucket.delete()
except google.api_core.exceptions.NotFound as err:
    print(f"error: bucket '{bucket_name}' not found: {err}", file=sys.stderr)
