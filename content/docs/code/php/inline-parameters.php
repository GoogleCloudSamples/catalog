function example_snippet(string $projectId): void
{
    $response = $client->translateText(
        ['Hello, world!'],
        'es',
        "projects/$projectId/locations/global"
    );
}
