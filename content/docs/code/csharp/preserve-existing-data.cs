public void UpdateKeyLabels(CryptoKeyName keyName)
{
    var client = KeyManagementServiceClient.Create();
    var cryptoKey = new CryptoKey
    {
        CryptoKeyName = keyName,
        Labels = { ["environment"] = "production" },
    };
    // List only modified fields so existing settings are preserved.
    var updateMask = new FieldMask { Paths = { "labels" } };

    CryptoKey updated = client.UpdateCryptoKey(cryptoKey, updateMask);
    Console.WriteLine($"Updated labels on {updated.Name}");
}
