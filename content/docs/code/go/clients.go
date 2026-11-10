ctx := context.Background()
c, err := foo.NewClient(ctx)
if err != nil {
	return fmt.Errorf("foo.NewClient: %w", err)
}
defer c.Close()
