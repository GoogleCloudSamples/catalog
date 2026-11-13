try
{
    client.DeleteBucket(bucketName);
}
catch (Google.GoogleApiException e) when (e.HttpStatusCode == System.Net.HttpStatusCode.NotFound)
{
    Console.Error.WriteLine($"error: bucket '{bucketName}' not found: {e.Message}");
}
