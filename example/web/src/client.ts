import { ApolloClient, InMemoryCache, HttpLink } from '@apollo/client/core/index.js';
import { createPersistedQueryLink } from '@apollo/client/link/persisted-queries/index.js';
export function createClient(uri = '/graphql') {
  return new ApolloClient({
    cache: new InMemoryCache(),
    link: createPersistedQueryLink({
      sha256: async (text) => {
        const digest = await globalThis.crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));
        return Array.from(new Uint8Array(digest), b => b.toString(16).padStart(2, '0')).join('');
      },
      retry: () => false,
    }).concat(new HttpLink({ uri })),
  });
}
