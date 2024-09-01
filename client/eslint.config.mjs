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
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default tseslint.config(
  {
    ignores: ["dist/**/*"],
  },

  // @ts-ignore
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
      // @ts-ignore
      react: fixupPluginRules(react),
      // @ts-ignore
      "react-hooks": fixupPluginRules(reactHooks),
      // @ts-ignore
      "jsx-a11y": fixupPluginRules(jsxA11Y),
    },
    settings: {
      react: {
        version: "detect",
      },
    },
    rules: {
      "no-restricted-globals": ["error", ...confusingBrowserGlobals],
    },
  },

  ...tseslint.configs.recommended.map((config) => ({
    files: ["**/*.ts?(x)"],
    ...config,
  })),
  {
    files: ["**/*.ts?(x)"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
    },
  },
);
