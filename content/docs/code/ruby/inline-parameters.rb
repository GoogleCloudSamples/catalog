def example_snippet project_id:
  response = client.translate_text(
    parent: "projects/#{project_id}/locations/global",
    contents: ["Hello, world!"],
    target_language_code: "es"
  )
end
