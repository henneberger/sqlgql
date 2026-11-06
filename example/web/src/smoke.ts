import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { print } from 'graphql';
import { createClient } from './client';
import { CustomerDocument, CustomersDocument, TypeDemoDocument, NumericDemoDocument, CustomerPageDocument, type CustomerPageQueryVariables } from './generated';
const client = createClient(process.env.API_URL ?? 'http://127.0.0.1:8080/graphql');
const query = (id: number, limit: number) => client.query({
 query: CustomerDocument,
 variables: {Query_customer_id: id, Query_customer_customerItems_limit: limit},
 fetchPolicy: 'network-only',
});
try {
 const result = await query(1, 10);
 assert.equal(result.data.customer?.name, 'Ada Lovelace');
 assert.equal(result.data.customer?.customerItems.length, 2);
 assert.equal((await query(2, 10)).data.customer, null, 'other tenant is isolated');
 await assert.rejects(query(1, 101), /forbidden/, 'OPA must reject excessive limits');
 const response = await fetch(process.env.API_URL ?? 'http://127.0.0.1:8080/graphql', {
  method: 'POST', headers: {'content-type':'application/json'},
  body: JSON.stringify({query:'{ customer(id: 1) { name } }'}),
 });
 assert.equal(response.status, 400, 'unregistered operation must be rejected');
 for (const limit of [1, 2]) {
  const list = await client.query({query: CustomersDocument, variables: {
   Query_customers_limit: 10, Query_customers_items_limit: limit,
  }, fetchPolicy: 'network-only'});
  assert.deepEqual(list.data.customers.map(c => c.id), [1, 3, 4]);
  assert.deepEqual(list.data.customers.map(c => c.items.length), [limit, limit, 0]);
  assert.deepEqual(list.data.customers[0].items.map(i => i.id), limit === 1 ? [101] : [101, 102]);
  assert.deepEqual(list.data.customers[1].items.map(i => i.id), limit === 1 ? [301] : [301, 302]);
  assert.equal(list.data.customers[1].email, null);
 }
 const uuid = '123e4567-e89b-12d3-a456-426614174000';
 const nilUUID = '00000000-0000-0000-0000-000000000000';
 const variables = {
  Query_typeDemo_id: uuid,
  Query_typeDemo_document: {nested: [true, null, 3], name: 'JSONB'},
  Query_typeDemo_raw: 'plain JSON string',
  Query_typeDemo_ids: [uuid, null, nilUUID],
  Query_typeDemo_labels: ['one', null, 'two'],
  Query_typeDemo_counts: [0, null, -1, 2147483647],
  Query_typeDemo_flags: [true, null, false],
  Query_typeDemo_scores: [1.5, null, -0.25],
  Query_typeDemo_documents: [{x: 1}, ['nested'], 'text', true, 42, null],
 };
 const roundtrip = await client.query({query: TypeDemoDocument, variables, fetchPolicy: 'network-only'});
 const typed = roundtrip.data.typeDemo!;
 assert.equal(typed.id, uuid);
 assert.deepEqual(typed.document, variables.Query_typeDemo_document);
 assert.equal(typed.raw, variables.Query_typeDemo_raw);
 assert.deepEqual(typed.ids, variables.Query_typeDemo_ids);
 assert.deepEqual(typed.labels, variables.Query_typeDemo_labels);
 assert.deepEqual(typed.counts, variables.Query_typeDemo_counts);
 assert.deepEqual(typed.flags, variables.Query_typeDemo_flags);
 assert.deepEqual(typed.scores, variables.Query_typeDemo_scores);
 assert.deepEqual(typed.documents, variables.Query_typeDemo_documents);
 assert.equal(typed.optionalId, null);
 assert.equal(typed.optionalDocument, null);
 assert.equal(typed.optionalLabels, null);
 assert.deepEqual(typed.emptyCounts, []);
 const empty = await client.query({query: TypeDemoDocument, variables: {...variables,
  Query_typeDemo_id: nilUUID, Query_typeDemo_document: 7, Query_typeDemo_raw: false,
  Query_typeDemo_ids: [], Query_typeDemo_labels: [], Query_typeDemo_counts: [],
  Query_typeDemo_flags: [], Query_typeDemo_scores: [], Query_typeDemo_documents: [],
 }, fetchPolicy: 'network-only'});
 assert.equal(empty.data.typeDemo?.id, nilUUID);
 assert.equal(empty.data.typeDemo?.document, 7);
 assert.equal(empty.data.typeDemo?.raw, false);
 for (const key of ['ids','labels','counts','flags','scores','documents'] as const) assert.deepEqual(empty.data.typeDemo?.[key], []);
 await assert.rejects(client.query({query: TypeDemoDocument, variables: {...variables, Query_typeDemo_id: 'invalid'}, fetchPolicy:'network-only'}), /UUID|uuid/);
 await assert.rejects(client.query({query: TypeDemoDocument, variables: {...variables, Query_typeDemo_ids: ['invalid']}, fetchPolicy:'network-only'}), /UUID|uuid/);
 const minInt64 = '-9223372036854775808';
 const maxInt64 = '9223372036854775807';
 const maxFloat32 = 3.4028234663852886e38;
 const minFloat32 = 1.401298464324817e-45;
 const numericVariables = (big: string, real: number, double: number, special = 'finite') => ({
  Query_numericDemo_big: big, Query_numericDemo_real: real, Query_numericDemo_double: double,
  Query_numericDemo_bigs: [minInt64, null, maxInt64, '9007199254740993'],
  Query_numericDemo_reals: [real, null, -real], Query_numericDemo_doubles: [double, null, -double],
  Query_numericDemo_special: special,
 });
 const numericCases: [string, number, number][] = [
  [minInt64, maxFloat32, Number.MAX_VALUE], [maxInt64, -maxFloat32, -Number.MAX_VALUE],
  ['9007199254740993', minFloat32, Number.MIN_VALUE], ['-9007199254740993', -minFloat32, -Number.MIN_VALUE],
  ['0', 16777217, 9007199254740992], ['1', 0.1, 0.1],
 ];
 for (const [big, real, double] of numericCases) {
  const numeric = await client.query({query: NumericDemoDocument, variables: numericVariables(big, real, double), fetchPolicy:'network-only'});
  const value = numeric.data.numericDemo!;
  assert.equal(value.big, big, 'bigint must survive JavaScript without rounding');
  assert.equal(value.real, Math.fround(real), 'float4 must round only to its native precision');
  assert.equal(value.double, double, 'float8 must preserve double precision');
  assert.deepEqual(value.bigs, [minInt64, null, maxInt64, '9007199254740993']);
  assert.deepEqual(value.reals, [Math.fround(real), null, Math.fround(-real)]);
  assert.deepEqual(value.doubles, [double, null, -double]);
  assert.equal(value.optionalBig, null); assert.equal(value.optionalReal, null); assert.equal(value.optionalDouble, null);
 }
 for (const big of ['9223372036854775808', '-9223372036854775809']) {
  await assert.rejects(client.query({query:NumericDemoDocument,variables:numericVariables(big,1,1),fetchPolicy:'network-only'}), /BigInt/);
 }
 await assert.rejects(client.query({query:NumericDemoDocument,variables:numericVariables(9007199254740992 as unknown as string,1,1),fetchPolicy:'network-only'}), /BigInt/);
 for (const real of [3.5e38, 1e-46]) {
  await assert.rejects(client.query({query:NumericDemoDocument,variables:numericVariables('1',real,1),fetchPolicy:'network-only'}), /float32|float4|underflow|range|unmarshal/);
 }
 // JavaScript cannot represent an overflowing float64; send its JSON token
 // directly using the same generated persisted document.
 const numericHash = createHash('sha256').update(print(NumericDemoDocument)).digest('hex');
 for (const token of ['1.8e308', '-1.8e308']) {
  const body = JSON.stringify({variables:numericVariables('1',1,1),extensions:{persistedQuery:{version:1,sha256Hash:numericHash}}})
   .replace('"Query_numericDemo_double":1', '"Query_numericDemo_double":' + token);
  const overflow = await fetch(process.env.API_URL ?? 'http://127.0.0.1:8080/graphql', {method:'POST',headers:{'content-type':'application/json'},body});
  const rejected = await overflow.json();
  assert.ok(rejected.errors?.some((error: {message: string}) => /range|Float/.test(error.message)), 'float8 overflow must be rejected');
 }
 for (const special of ['nan','infinity','-infinity']) {
  await assert.rejects(client.query({query:NumericDemoDocument,variables:numericVariables('1',1,1,special),fetchPolicy:'network-only'}), /NaN|infinite|marshal/);
 }
 const page = async (variables: CustomerPageQueryVariables = {}) => (await client.query({query:CustomerPageDocument,variables,fetchPolicy:'network-only'})).data.customerPage;
 const firstPage = await page({Query_customerPage_first:1,Query_customerPage_pageItems_first:1});
 assert.deepEqual(firstPage.edges.map(e => e.node.id), [1]);
 assert.equal(firstPage.pageInfo.hasNextPage, true);
 assert.equal(firstPage.pageInfo.hasPreviousPage, false);
 assert.equal(firstPage.pageInfo.startCursor, firstPage.edges[0].cursor);
 assert.equal(firstPage.pageInfo.endCursor, firstPage.edges[0].cursor);
 const firstCursor = firstPage.pageInfo.endCursor!;
 const childCursor = firstPage.edges[0].node.pageItems.pageInfo.endCursor!;
 assert.deepEqual(firstPage.edges[0].node.pageItems.edges.map(e => e.node.id), [101]);
 assert.equal(firstPage.edges[0].node.pageItems.pageInfo.hasNextPage, true);
 const childNext = await page({Query_customerPage_first:1,Query_customerPage_pageItems_first:1,Query_customerPage_pageItems_after:childCursor});
 assert.deepEqual(childNext.edges[0].node.pageItems.edges.map(e => e.node.id), [102]);
 assert.equal(childNext.edges[0].node.pageItems.pageInfo.hasNextPage, false);
 const secondPage = await page({Query_customerPage_first:1,Query_customerPage_after:firstCursor});
 assert.deepEqual(secondPage.edges.map(e => e.node.id), [3]);
 const secondCursor = secondPage.pageInfo.endCursor!;
 const thirdPage = await page({Query_customerPage_first:1,Query_customerPage_after:secondCursor});
 assert.deepEqual(thirdPage.edges.map(e => e.node.id), [4]);
 assert.equal(thirdPage.pageInfo.hasNextPage, false);
 assert.deepEqual(thirdPage.edges[0].node.pageItems.edges, []);
 const thirdCursor = thirdPage.pageInfo.endCursor!;
 const pastEnd = await page({Query_customerPage_first:2,Query_customerPage_after:thirdCursor});
 assert.deepEqual(pastEnd.edges, []);
 assert.equal(pastEnd.pageInfo.startCursor, null); assert.equal(pastEnd.pageInfo.endCursor, null);
 const backward = await page({Query_customerPage_last:2});
 assert.deepEqual(backward.edges.map(e => e.node.id), [3,4], 'backward output stays in canonical ascending order');
 assert.equal(backward.pageInfo.hasPreviousPage, true);
 const backwardBefore = await page({Query_customerPage_last:1,Query_customerPage_before:thirdCursor});
 assert.deepEqual(backwardBefore.edges.map(e => e.node.id), [3]);
 assert.equal(backwardBefore.pageInfo.hasPreviousPage, true);
 const backwardStart = await page({Query_customerPage_last:1,Query_customerPage_before:firstCursor});
 assert.deepEqual(backwardStart.edges, []); assert.equal(backwardStart.pageInfo.hasPreviousPage, false);
 const bounded = await page({Query_customerPage_first:10,Query_customerPage_after:firstCursor,Query_customerPage_before:thirdCursor});
 assert.deepEqual(bounded.edges.map(e => e.node.id), [3]);
 const zero = await page({Query_customerPage_first:0});
 assert.deepEqual(zero.edges, []); assert.equal(zero.pageInfo.hasNextPage, true); assert.equal(zero.pageInfo.startCursor,null);
 const all = await page();
 assert.deepEqual(all.edges.map(e => e.node.id), [1,3,4]);
 assert.equal(all.pageInfo.hasNextPage, false);
 const backwardChild = await page({Query_customerPage_first:1,Query_customerPage_pageItems_last:1});
 assert.deepEqual(backwardChild.edges[0].node.pageItems.edges.map(e => e.node.id), [102]);
 assert.equal(backwardChild.edges[0].node.pageItems.pageInfo.hasPreviousPage,true);
 for (const invalid of [
  {Query_customerPage_first:-1}, {Query_customerPage_first:101}, {Query_customerPage_first:1,Query_customerPage_last:1},
  {Query_customerPage_after:'not a cursor'}, {Query_customerPage_after:''}, {Query_customerPage_after:childCursor},
  {Query_customerPage_first:1,Query_customerPage_after:firstCursor,Query_customerPage_pageItems_after:childCursor},
 ]) await assert.rejects(page(invalid), /cursor|page size|first or last/);
 const wrongType = JSON.parse(Buffer.from(firstCursor,'base64url').toString()); wrongType.k[0]='not an integer';
 await assert.rejects(page({Query_customerPage_after:Buffer.from(JSON.stringify(wrongType)).toString('base64url')}), /unmarshal/);
 console.log('PASS: typed Apollo persisted query, nested retrieval, tenant isolation, OPA denial, operation allowlist, per-parent batched limits, UUID/JSON/array round-trips, bigint/float/double boundaries, opt-in Relay forward/backward pagination');
} finally {client.stop();}
