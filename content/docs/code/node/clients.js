import {v2} from '@google-cloud/dlp';

async function listInfoTypes() {
  const dlpClient = new v2.DlpServiceClient();
  try {
    // Snippet content ...
  } finally {
    await dlpClient.close();
  }
}

