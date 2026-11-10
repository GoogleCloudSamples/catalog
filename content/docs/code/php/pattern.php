use Google\Cloud\Dlp\V2\Client\DlpServiceClient;
use Google\Cloud\Dlp\V2\ListInfoTypesRequest;

/**
 * Prints the sensitive data detectors (infoTypes) supported by Cloud Data Loss
 * Prevention (DLP) for risk analysis.
 *
 * InfoTypes represent categories of sensitive data—such as person names, email
 * addresses, or credit card numbers—that DLP uses to scan and redact content.
 */
function list_info_types(): void
{
    $dlp = new DlpServiceClient();

    $request = (new ListInfoTypesRequest())
        // Return infoTypes that risk analysis jobs can detect.
        ->setFilter('supported_by=RISK_ANALYSIS');

    $response = $dlp->listInfoTypes($request);

    print('Info types found:' . PHP_EOL);
    foreach ($response->getInfoTypes() as $infoType) {
        printf('Name: %s' . PHP_EOL, $infoType->getName());
        printf('Display name: %s' . PHP_EOL, $infoType->getDisplayName());
    }
}
