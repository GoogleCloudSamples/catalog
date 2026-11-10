// [START sandwichmanager_make_sandwich]
import {v1} from '@example-cloud/sandwichmanager';

async function makeSandwich() {
  const client = new v1.SandwichManagerClient();
  try {
    await client.makeSandwich();
  } finally {
    await client.close();
  }
}
// [END sandwichmanager_make_sandwich]
