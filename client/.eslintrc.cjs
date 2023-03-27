module.exports = {
  plugins: ["react", "jsx-a11y"],
  extends: [
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:jsx-a11y/recommended",
  ],
  ignorePatterns: ["dist/**"],
  env: { node: true },
  settings: {
    react: { version: "detect" },
  },
  overrides: [
    {
      files: ["src/**"],
      env: { browser: true },
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
