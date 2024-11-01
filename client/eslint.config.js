// @ts-check

import path from "node:path";
import { fileURLToPath } from "node:url";

import confusingBrowserGlobals from "confusing-browser-globals";

import js from "@eslint/js";
import { fixupConfigRules, fixupPluginRules } from "@eslint/compat";
import { FlatCompat } from "@eslint/eslintrc";

import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import jsxA11Y from "eslint-plugin-jsx-a11y";
import tseslint from "typescript-eslint";

const compat = new FlatCompat({
  baseDirectory: path.dirname(fileURLToPath(import.meta.url)),
});

export default tseslint.config(
  {
    ignores: ["dist/**/*"],
  },

  js.configs.recommended,

  // @ts-expect-error
  react.configs.flat.recommended,
  { settings: { react: { version: "detect" } } },

  {
    plugins: { "react-hooks": reactHooks },
    rules: reactHooks.configs.recommended.rules,
  },

  ...tseslint.configs.recommended.map((config) => ({
    files: ["**/*.ts?(x)"],
    ...config,
  })),

  { plugins: { "jsx-a11y": fixupPluginRules(jsxA11Y) } },
  ...fixupConfigRules(compat.extends("plugin:jsx-a11y/recommended")),

  {
    rules: {
      "no-restricted-globals": ["error", ...confusingBrowserGlobals],
    },
  },
);
