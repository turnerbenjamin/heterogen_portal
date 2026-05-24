import typescriptEslintParser from "@typescript-eslint/parser";
import typescriptEslintPlugin from "@typescript-eslint/eslint-plugin";

export default [
  {
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parser: typescriptEslintParser,
      parserOptions: {
        project: "./tsconfig.json",
        ecmaVersion: 2020,
        sourceType: "module",
      },
      globals: {
        browser: true,
        commonjs: true,
        es6: true,
        node: true,
        window: "readonly",
        Xrm: "readonly",
      },
    },
    plugins: {
      "@typescript-eslint": typescriptEslintPlugin,
    },
    rules: {
      "no-unused-vars": "warn",
      eqeqeq: "error",
      "linebreak-style": "warn",
      "@typescript-eslint/no-explicit-any": "warn",
      "no-console": "warn",
      "no-global-assign": "warn",
      semi: ["warn", "always"],
    },
  },
];
