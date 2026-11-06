import { defineConfig } from 'vite';
export default defineConfig({server:{proxy:{'/graphql':'http://127.0.0.1:8080'}}});
