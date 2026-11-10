try {
  await storage.bucket(bucketName).delete();
} catch (err) {
  console.error(`error: failed to delete bucket '${bucketName}': ${err.message}`);
}
