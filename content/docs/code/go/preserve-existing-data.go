func updateKeyLabels(w io.Writer, keyName string) error {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("kms.NewKeyManagementClient: %w", err)
	}
	defer client.Close()

	key, err := client.UpdateCryptoKey(ctx, &kmspb.UpdateCryptoKeyRequest{
		CryptoKey: &kmspb.CryptoKey{
			Name:   keyName,
			Labels: map[string]string{"environment": "production"},
		},
		// List only modified fields so existing settings are preserved.
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"labels"},
		},
	})
	if err != nil {
		return fmt.Errorf("UpdateCryptoKey: %w", err)
	}
	fmt.Fprintf(w, "Updated labels on %s\n", key.GetName())
	return nil
}
