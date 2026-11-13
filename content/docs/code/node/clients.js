import {v2} from '@google-cloud/dlp';

async function listInfoTypes() {
  const client = new v2.DlpServiceClient();
  try {
    // Snippet content ...
  } finally {
    await client.close();
  }
}

