begin
  bucket = client.bucket bucket_name
  bucket.delete
rescue Google::Cloud::NotFoundError => e
  warn "error: bucket '#{bucket_name}' not found: #{e.message}"
end
