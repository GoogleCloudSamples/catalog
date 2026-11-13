try {
    $bucket = $client->bucket($bucketName);
    $bucket->delete();
} catch (NotFoundException $e) {
    fwrite(STDERR, sprintf("error: bucket '%s' not found: %s" . PHP_EOL, $bucketName, $e->getMessage()));
}
