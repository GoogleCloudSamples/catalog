// Prints the sensitive data detectors (infoTypes) supported by Cloud Data Loss
// Prevention (DLP) for risk analysis.
//
// InfoTypes represent categories of sensitive data—such as person names, email
// addresses, or credit card numbers—that DLP uses to scan and redact content.
public static void listInfoTypes() throws IOException {
  try (DlpServiceClient dlpClient = DlpServiceClient.create()) {
    ListInfoTypesRequest request =
        ListInfoTypesRequest.newBuilder()
            // Return infoTypes that risk analysis jobs can detect.
            .setFilter("supported_by=RISK_ANALYSIS")
            .build();

    ListInfoTypesResponse response = dlpClient.listInfoTypes(request);

    System.out.println("Info types found:");
    for (InfoTypeDescription infoType : response.getInfoTypesList()) {
      System.out.println("Name: " + infoType.getName());
      System.out.println("Display name: " + infoType.getDisplayName());
    }
  }
}
