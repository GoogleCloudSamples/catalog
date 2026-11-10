def encrypt key_name:
  client = Google::Cloud::Kms::V1::KeyManagementService::Client.new
  response = client.encrypt name: key_name, plaintext: "Hello, world!"

  # Use sensitive data without printing it.
  ciphertext = response.ciphertext
  puts "Ciphertext size: #{ciphertext.bytesize} bytes"
end
