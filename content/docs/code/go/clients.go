ctx := context.Background()
client, err := foo.NewClient(ctx)
if err != nil {
	return fmt.Errorf("foo.NewClient: %w", err)
}
defer client.Close()
