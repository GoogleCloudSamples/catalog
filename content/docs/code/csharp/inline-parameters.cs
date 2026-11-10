public void Example(string projectId = "example-project-id")
{
    var response = client.TranslateText(new TranslateTextRequest
    {
        Parent = $"projects/{projectId}/locations/global",
        TargetLanguageCode = "es",
        Contents = { "Hello, world!" }
    });
}
