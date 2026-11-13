try {
  await client.bucket(bucketName).delete();
} catch (err) {
  console.error(`error: failed to delete bucket '${bucketName}': ${err.message}`);
}
