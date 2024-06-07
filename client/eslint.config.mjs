import path from "node:path";
import { fileURLToPath } from "node:url";

import confusingBrowserGlobals from "confusing-browser-globals";

import js from "@eslint/js";
import { fixupConfigRules, fixupPluginRules } from "@eslint/compat";
import { FlatCompat } from "@eslint/eslintrc";

import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import jsxA11Y from "eslint-plugin-jsx-a11y";
import typescriptEslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";

const compat = new FlatCompat({
  baseDirectory: path.dirname(fileURLToPath(import.meta.url)),
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [
  {
    ignores: ["dist/**/*"],
  },

  ...fixupConfigRules(
    compat.extends(
      "eslint:recommended",
      "plugin:react/recommended",
      "plugin:react-hooks/recommended",
      "plugin:jsx-a11y/recommended",
    ),
  ),
  {
    plugins: {
      react: fixupPluginRules(react),
      "react-hooks": fixupPluginRules(reactHooks),
      "jsx-a11y": fixupPluginRules(jsxA11Y),
    },
    settings: {
      react: {
        version: "detect",
      },
    },
    rules: {
      "no-restricted-globals": ["error"].concat(confusingBrowserGlobals),
    },
  },

  ...compat.extends("plugin:@typescript-eslint/recommended").map((config) => ({
    files: ["**/*.ts?(x)"],
    ...config,
  })),
  {
    files: ["**/*.ts?(x)"],
    plugins: {
      "@typescript-eslint": typescriptEslint,
    },
    languageOptions: {
      parser: tsParser,
      ecmaVersion: 5,
      sourceType: "script",
      parserOptions: {
        project: ["./tsconfig.json", "./tsconfig.node.json"],
      },
    },
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
    },
  },
];
