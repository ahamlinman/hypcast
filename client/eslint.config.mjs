import { fixupConfigRules, fixupPluginRules } from "@eslint/compat";
import jsxA11Y from "eslint-plugin-jsx-a11y";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";
import typescriptEslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
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
      "plugin:jsx-a11y/recommended",
      "plugin:react/recommended",
      "plugin:react-hooks/recommended",
    ),
  ),

  {
    plugins: {
      "jsx-a11y": fixupPluginRules(jsxA11Y),
      react: fixupPluginRules(react),
      "react-hooks": fixupPluginRules(reactHooks),
    },
    languageOptions: {
      globals: {
        ...globals.node,
      },
    },
    settings: {
      react: {
        version: "detect",
      },
    },
    rules: {
      "no-restricted-globals": [
        "error",
        "addEventListener",
        "blur",
        "close",
        "closed",
        "confirm",
        "defaultStatus",
        "defaultstatus",
        "event",
        "external",
        "find",
        "focus",
        "frameElement",
        "frames",
        "history",
        "innerHeight",
        "innerWidth",
        "length",
        "location",
        "locationbar",
        "menubar",
        "moveBy",
        "moveTo",
        "name",
        "onblur",
        "onerror",
        "onfocus",
        "onload",
        "onresize",
        "onunload",
        "open",
        "opener",
        "opera",
        "outerHeight",
        "outerWidth",
        "pageXOffset",
        "pageYOffset",
        "parent",
        "print",
        "removeEventListener",
        "resizeBy",
        "resizeTo",
        "screen",
        "screenLeft",
        "screenTop",
        "screenX",
        "screenY",
        "scroll",
        "scrollbars",
        "scrollBy",
        "scrollTo",
        "scrollX",
        "scrollY",
        "self",
        "status",
        "statusbar",
        "stop",
        "toolbar",
        "top",
      ],
    },
  },

  {
    files: ["src/**"],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...Object.fromEntries(
          Object.entries(globals.node).map(([key]) => [key, "off"]),
        ),
      },
    },
  },

  ...compat.extends("plugin:@typescript-eslint/recommended").map((config) => ({
    ...config,
    files: ["**/*.ts?(x)"],
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
