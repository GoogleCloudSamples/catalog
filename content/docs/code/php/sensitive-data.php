function encrypt(string $keyName): void
{
    $client = new KeyManagementServiceClient();
    $request = (new EncryptRequest())->setName($keyName)->setPlaintext('Hello, world!');
    $response = $client->encrypt($request);

    // Use sensitive data without printing it.
    $ciphertext = $response->getCiphertext();
    printf('Ciphertext size: %d bytes' . PHP_EOL, strlen($ciphertext));
}
