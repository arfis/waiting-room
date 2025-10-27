import nx from '@nx/eslint-plugin';

export default [
  ...nx.configs['flat/base'],
  ...nx.configs['flat/typescript'],
  ...nx.configs['flat/javascript'],
  {
    ignores: ['**/dist'],
  },
  {
    files: ['**/*.ts', '**/*.tsx', '**/*.js', '**/*.jsx'],
    rules: {
      '@nx/enforce-module-boundaries': [
        'error',
        {
          enforceBuildableLibDependency: true,
          allow: ['^.*/eslint(\\.base)?\\.config\\.[cm]?[jt]s$'],
          depConstraints: [
            {
              sourceTag: 'app:admin',
              onlyDependOnLibsWithTags: ['lib:api-client', 'lib:ui', 'lib:primeng-components'],
            },
            {
              sourceTag: 'app:backoffice',
              onlyDependOnLibsWithTags: ['lib:api-client', 'lib:ui', 'lib:primeng-components'],
            },
            {
              sourceTag: 'app:kiosk',
              onlyDependOnLibsWithTags: ['lib:api-client', 'lib:ui', 'lib:primeng-components'],
            },
            {
              sourceTag: 'app:mobile',
              onlyDependOnLibsWithTags: ['lib:api-client', 'lib:ui', 'lib:primeng-components'],
            },
            {
              sourceTag: 'app:tv',
              onlyDependOnLibsWithTags: ['lib:api-client', 'lib:ui', 'lib:primeng-components'],
            },
            {
              sourceTag: 'lib:api-client',
              onlyDependOnLibsWithTags: [],
            },
            {
              sourceTag: 'lib:ui',
              onlyDependOnLibsWithTags: [],
            },
            {
              sourceTag: 'lib:primeng-components',
              onlyDependOnLibsWithTags: [],
            },
          ],
        },
      ],
    },
  },
  {
    files: [
      '**/*.ts',
      '**/*.tsx',
      '**/*.cts',
      '**/*.mts',
      '**/*.js',
      '**/*.jsx',
      '**/*.cjs',
      '**/*.mjs',
    ],
    // Override or add rules here
    rules: {},
  },
];
