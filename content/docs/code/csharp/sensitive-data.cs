public void Encrypt(CryptoKeyName keyName)
{
    var client = KeyManagementServiceClient.Create();
    var response = client.Encrypt(keyName, ByteString.CopyFromUtf8("Hello, world!"));

    // Use sensitive data without printing it.
    ByteString ciphertext = response.Ciphertext;
    Console.WriteLine($"Ciphertext size: {ciphertext.Length} bytes");
}
