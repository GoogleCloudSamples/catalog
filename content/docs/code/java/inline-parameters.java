public static void exampleSnippet(String projectId) {
    TranslateTextRequest request =
        TranslateTextRequest.newBuilder()
            .setParent(String.format("projects/%s/locations/global", projectId))
            .setTargetLanguageCode("es")
            .addContents("Hello, world!")
            .build();

    client.translateText(request);
}
