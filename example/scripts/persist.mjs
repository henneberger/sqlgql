// Standard GraphQL print + SHA-256, shared with Apollo's persisted-query link.
import { createRequire } from 'node:module';
import { createHash } from 'node:crypto';
import { readFile, readdir, writeFile } from 'node:fs/promises';
const require = createRequire(new URL('../web/package.json', import.meta.url));
const { parse, print } = require('graphql');
const directory = new URL('../generated/operations/', import.meta.url);
const manifest = {};
for (const file of (await readdir(directory)).filter(f => f.endsWith('.graphql')).sort()) {
  const query = print(parse(await readFile(new URL(file, directory), 'utf8')));
  manifest[createHash('sha256').update(query).digest('hex')] = query;
}
await writeFile(new URL('../generated/persisted.json', import.meta.url), JSON.stringify(manifest, null, 2) + '\n');
