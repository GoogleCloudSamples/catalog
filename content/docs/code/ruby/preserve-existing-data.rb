def update_key_labels key_name:
  client = Google::Cloud::Kms::V1::KeyManagementService::Client.new
  key = client.update_crypto_key(
    crypto_key:  {
      name:   key_name,
      labels: { "environment" => "production" }
    },
    # List only modified fields so existing settings are preserved.
    update_mask: { paths: ["labels"] }
  )
  puts "Updated labels on #{key.name}"
end
