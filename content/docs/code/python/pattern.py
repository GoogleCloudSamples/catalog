def list_info_types() -> None:
    """Prints the sensitive data detectors (infoTypes) supported by Cloud Data
    Loss Prevention (DLP) for risk analysis.

    InfoTypes represent categories of sensitive data—such as person names,
    email addresses, or credit card numbers—that DLP uses to scan and redact
    content.
    """
    with dlp_v2.DlpServiceClient() as dlp_client:
        request = dlp_v2.ListInfoTypesRequest(
            # Return infoTypes that risk analysis jobs can detect.
            filter="supported_by=RISK_ANALYSIS",
        )

        response = dlp_client.list_info_types(request=request)

        print("Info types found:")
        for info_type in response.info_types:
            print(f"Name: {info_type.name}")
            print(f"Display name: {info_type.display_name}")
