def example_snippet(project_id: str) -> None:
    response = client.translate_text(
        parent=f"projects/{project_id}/locations/global",
        contents=["Hello, world!"],
        target_language_code="es",
    )
