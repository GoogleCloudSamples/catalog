func makeSandwich() error {
	ctx := context.Background()
	client, err := sandwichmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("sandwichmanager.NewClient: %w", err)
	}
	defer client.Close()

	if err := client.MakeSandwich(ctx); err != nil {
		return fmt.Errorf("MakeSandwich: %w", err)
	}

	return nil
}
