const path = require('path');
const autoprefixer = require('autoprefixer');
const tailwindcss = require('tailwindcss');

module.exports = {
  plugins: [
    tailwindcss({
      config: path.resolve(__dirname, 'tailwind.config.cjs')
    }),
    autoprefixer()
  ]
};
