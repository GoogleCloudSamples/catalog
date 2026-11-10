try {
  storage.delete(bucketName, blobName);
} catch (StorageException e) {
  System.err.printf(
      "error: failed to delete '%s' in bucket '%s': %s%n",
      blobName, bucketName, e.getMessage());
}
