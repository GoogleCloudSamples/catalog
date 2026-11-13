# Prints the sensitive data detectors (infoTypes) supported by Cloud Data Loss
# Prevention (DLP) for risk analysis.
#
# InfoTypes represent categories of sensitive data—such as person names, email
# addresses, or credit card numbers—that DLP uses to scan and redact content.
def list_info_types
  client = Google::Cloud::Dlp::V2::DlpService::Client.new

  request = {
    # Return infoTypes that risk analysis jobs can detect.
    filter: "supported_by=RISK_ANALYSIS"
  }

  response = client.list_info_types request

  puts "Info types found:"
  response.info_types.each do |info_type|
    puts "Name: #{info_type.name}"
    puts "Display name: #{info_type.display_name}"
  end
end
