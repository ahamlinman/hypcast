module.exports = {
    "env": {
        "browser": true,
        "node": true
    },
    "extends": "eslint:recommended",
    "parserOptions": {
        "ecmaVersion": 2017,
        "ecmaFeatures": {
            "jsx": true
        },
        "sourceType": "module"
    },
    "plugins": [
        "react"
    ],
    "rules": {
        "comma-dangle": [ "error", "always-multiline" ],
        "comma-spacing": [ "error", { "before": false, "after": true } ],
        "comma-style": [ "error", "last" ],
        "eqeqeq": [ "error", "always" ],
        "indent": [ "error", 2 ],
        "linebreak-style": [ "error", "unix" ],
        "no-tabs": [ "error" ],
        "no-trailing-spaces": [ "error" ],
        "no-unneeded-ternary": [ "error" ],
        "quotes": [ "error", "single" ],
        "semi": [ "error", "always" ],
        "semi-spacing": [ "error", { "before": false, "after": true } ],

        "no-await-in-loop": [ "warn" ],
        "no-return-await": [ "warn" ],
        "require-await": [ "warn" ],

        "react/jsx-uses-react": [ "error" ],
        "react/jsx-uses-vars": [ "error" ],
        "react/react-in-jsx-scope": [ "error" ],
    }
};
