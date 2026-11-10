function update_key_labels(string $keyName): void
{
    $client = new KeyManagementServiceClient();
    $cryptoKey = (new CryptoKey())
        ->setName($keyName)
        ->setLabels(['environment' => 'production']);
    // List only modified fields so existing settings are preserved.
    $updateMask = (new FieldMask())->setPaths(['labels']);

    $request = (new UpdateCryptoKeyRequest())
        ->setCryptoKey($cryptoKey)
        ->setUpdateMask($updateMask);
    $key = $client->updateCryptoKey($request);
    printf('Updated labels on %s' . PHP_EOL, $key->getName());
}
