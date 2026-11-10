def update_key_labels(key_name: str) -> None:
    with kms_v1.KeyManagementServiceClient() as client:
        key = client.update_crypto_key(
            request={
                "crypto_key": {
                    "name": key_name,
                    "labels": {"environment": "production"},
                },
                # List only modified fields so existing settings are preserved.
                "update_mask": {"paths": ["labels"]},
            }
        )
        print(f"Updated labels on {key.name}")
