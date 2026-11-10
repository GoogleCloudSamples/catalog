def encrypt(key_name: str) -> None:
    with kms_v1.KeyManagementServiceClient() as client:
        response = client.encrypt(request={"name": key_name, "plaintext": b"Hello, world!"})

        # Use sensitive data without printing it.
        ciphertext = response.ciphertext
        print(f"Ciphertext size: {len(ciphertext)} bytes")
