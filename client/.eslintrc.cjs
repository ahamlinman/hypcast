const restrictedGlobals = require("confusing-browser-globals");

module.exports = {
  plugins: ["jsx-a11y", "react", "react-hooks"],
  extends: [
    "eslint:recommended",
    "plugin:jsx-a11y/recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended",
  ],
  ignorePatterns: ["dist/**"],
  env: {
    es6: true,
    node: true,
  },
  settings: {
    react: { version: "detect" },
  },
  rules: {
    "no-restricted-globals": ["error"].concat(restrictedGlobals),
  },
  overrides: [
    {
      files: ["src/**"],
      env: {
        browser: true,
        node: false,
      },
    },
    {
      files: ["**/*.ts?(x)"],
      extends: ["plugin:@typescript-eslint/recommended"],
      plugins: ["@typescript-eslint"],
      parser: "@typescript-eslint/parser",
      parserOptions: {
        project: ["./tsconfig.json", "./tsconfig.node.json"],
      },
      rules: {
        "@typescript-eslint/no-explicit-any": "off",
      },
    },
  ],
};
