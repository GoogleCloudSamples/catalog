public static void updateKeyLabels(CryptoKeyName keyName) throws IOException {
  try (KeyManagementServiceClient client = KeyManagementServiceClient.create()) {
    CryptoKey cryptoKey =
        CryptoKey.newBuilder()
            .setName(keyName.toString())
            .putLabels("environment", "production")
            .build();
    // List only modified fields so existing settings are preserved.
    FieldMask updateMask = FieldMask.newBuilder().addPaths("labels").build();

    CryptoKey updated = client.updateCryptoKey(cryptoKey, updateMask);
    System.out.println("Updated labels on " + updated.getName());
  }
}
