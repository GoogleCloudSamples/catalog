async function encrypt(keyName) {
  const client = new v1.KeyManagementServiceClient();
  const [response] = await client.encrypt({
    name: keyName,
    plaintext: Buffer.from('Hello, world!'),
  });

    // Use sensitive data without printing it.
    const ciphertext = response.ciphertext;
    console.log(`Ciphertext size: ${ciphertext.length} bytes`);
  } finally {
    await client.close();
  }
}
