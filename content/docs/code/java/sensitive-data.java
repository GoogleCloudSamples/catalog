public static void encrypt(CryptoKeyName keyName) throws IOException {
  try (KeyManagementServiceClient client = KeyManagementServiceClient.create()) {
    EncryptResponse response = client.encrypt(keyName, ByteString.copyFromUtf8("Hello, world!"));

    // Use sensitive data without printing it.
    ByteString ciphertext = response.getCiphertext();
    System.out.println("Ciphertext size: " + ciphertext.size() + " bytes");
  }
}
