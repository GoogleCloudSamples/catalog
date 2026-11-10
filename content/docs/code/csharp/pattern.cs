// Prints the sensitive data detectors (infoTypes) supported by Cloud Data Loss
// Prevention (DLP) for risk analysis.
//
// InfoTypes represent categories of sensitive data—such as person names, email
// addresses, or credit card numbers—that DLP uses to scan and redact content.
public void ListInfoTypes()
{
    var dlp = DlpServiceClient.Create();

    var request = new ListInfoTypesRequest
    {
        // Return infoTypes that risk analysis jobs can detect.
        Filter = "supported_by=RISK_ANALYSIS"
    };

    var response = dlp.ListInfoTypes(request);

    Console.WriteLine("Info types found:");
    foreach (var infoType in response.InfoTypes)
    {
        Console.WriteLine($"Name: {infoType.Name}");
        Console.WriteLine($"Display name: {infoType.DisplayName}");
    }
}
