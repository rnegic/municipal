export default {
  extends: ['stylelint-config-standard-scss'],
  rules: {
    // !important` запрещён
    'declaration-no-important': true,
    // цвета берём только из токенов, hex в компонентах запрещён.
    'color-no-hex': true,
    // классы внутри CSS Modules — camelCase, keyframes — тоже.
    'selector-class-pattern': null,
    'keyframes-name-pattern': null,
    // Модульные классы склеиваются в один элемент
    'no-descending-specificity': null,
  },
  overrides: [
    {
      // tokens.scss — единственное место, где допустимо объявлять значения
      files: ['src/app/styles/tokens.scss'],
      rules: {
        'color-no-hex': null,
      },
    },
  ],
}
