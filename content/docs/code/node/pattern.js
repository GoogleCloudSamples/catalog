import {v2} from '@google-cloud/dlp';

/**
 * Prints the sensitive data detectors (infoTypes) supported by Cloud Data Loss
 * Prevention (DLP) for risk analysis.
 *
 * InfoTypes represent categories of sensitive data—such as person names, email
 * addresses, or credit card numbers—that DLP uses to scan and redact content.
 */
async function listInfoTypes() {
  const client = new v2.DlpServiceClient();
  try {
    const [response] = await client.listInfoTypes({
      // Return infoTypes that risk analysis jobs can detect.
      filter: 'supported_by=RISK_ANALYSIS',
    });

    console.log('Info types found:');
    for (const infoType of response.infoTypes) {
      console.log(`Name: ${infoType.name}`);
      console.log(`Display name: ${infoType.displayName}`);
    }
  } finally {
    await client.close();
  }
}
