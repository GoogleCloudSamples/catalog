func encrypt(w io.Writer, keyName string) error {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("kms.NewKeyManagementClient: %w", err)
	}
	defer client.Close()

	resp, err := client.Encrypt(ctx, &kmspb.EncryptRequest{
		Name:      keyName,
		Plaintext: []byte("Hello, world!"),
	})
	if err != nil {
		return fmt.Errorf("Encrypt: %w", err)
	}

	// Use sensitive data without printing it.
	ciphertext := resp.GetCiphertext()
	fmt.Fprintf(w, "Ciphertext size: %d bytes\n", len(ciphertext))
	return nil
}
