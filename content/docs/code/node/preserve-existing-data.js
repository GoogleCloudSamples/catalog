async function updateKeyLabels(keyName) {
  const client = new v1.KeyManagementServiceClient();
  try {
    const [key] = await client.updateCryptoKey({
      cryptoKey: {
        name: keyName,
        labels: {environment: 'production'},
      },
      // List only modified fields so existing settings are preserved.
      updateMask: {paths: ['labels']},
    });
    console.log(`Updated labels on ${key.name}`);
  } finally {
    await client.close();
  }
}
