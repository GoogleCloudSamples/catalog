func makeSandwich() error {
	ctx := context.Background()
	c, err := sandwichmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("sandwichmanager.NewClient: %w", err)
	}
	defer c.Close()

	if err := c.MakeSandwich(ctx); err != nil {
		return fmt.Errorf("MakeSandwich: %w", err)
	}

	return nil
}
