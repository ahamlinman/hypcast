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

  react.configs.flat.recommended,
  { settings: { react: { version: "detect" } } },

  ...tseslint.configs.recommended.map((config) => ({
    files: ["**/*.ts?(x)"],
    ...config,
  })),

  ...fixupConfigRules(
    compat.extends(
      "plugin:react-hooks/recommended",
      "plugin:jsx-a11y/recommended",
    ),
  ),
  {
    plugins: {
      // @ts-ignore
      "react-hooks": fixupPluginRules(reactHooks),
      "jsx-a11y": fixupPluginRules(jsxA11Y),
    },
  },

  {
    rules: {
      "no-restricted-globals": ["error", ...confusingBrowserGlobals],
    },
  },
);
