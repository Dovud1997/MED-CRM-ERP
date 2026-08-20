import eslint from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
export default tseslint.config(
  { ignores: ['**/dist/**','**/.next/**','**/generated/**','**/coverage/**','**/next-env.d.ts','**/next.config.ts'] },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  { files: ['**/*.ts','**/*.tsx'], languageOptions: { globals: { ...globals.node, ...globals.browser }, parserOptions: { projectService: true, tsconfigRootDir: import.meta.dirname } }, rules: { '@typescript-eslint/no-explicit-any': 'error', '@typescript-eslint/no-unused-vars': ['error',{argsIgnorePattern:'^_'}] } }
);
