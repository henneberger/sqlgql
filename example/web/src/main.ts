import { createClient } from './client';
import { CustomerDocument } from './generated';
import './style.css';
const client = createClient();
const form = document.querySelector('form')!;
const output = document.querySelector('pre')!;
form.addEventListener('submit', async event => {
 event.preventDefault();
 output.textContent = 'Loading…';
 try {
  const result = await client.query({query: CustomerDocument, variables: {
   Query_customer_id: Number((document.querySelector('#customer') as HTMLInputElement).value),
   Query_customer_customerItems_limit: Number((document.querySelector('#limit') as HTMLInputElement).value),
  }, fetchPolicy: 'network-only'});
  output.textContent = JSON.stringify(result.data, null, 2);
 } catch(error) {output.textContent = error instanceof Error ? error.message : String(error);}
});
form.requestSubmit();
