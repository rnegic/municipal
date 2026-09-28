export default {
  extends: ['stylelint-config-standard-scss'],
  rules: {
    'declaration-no-important': true,
    'color-no-hex': true,
    'selector-class-pattern': null,
    'keyframes-name-pattern': null,
    'no-descending-specificity': null,
  },
  overrides: [
    {
      files: ['src/app/styles/tokens.scss'],
      rules: {
        'color-no-hex': null,
      },
    },
  ],
}
