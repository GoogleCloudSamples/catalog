function exampleSnippet(projectId) {
  const request = {
    parent: `projects/${projectId}/locations/global`,
    contents: ['Hello, world!'],
    targetLanguageCode: 'es',
  };

  client.translateText(request);
}
