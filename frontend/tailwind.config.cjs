/** @type {import('tailwindcss').Config} */
module.exports = {
  content: {
    relative: true,
    files: ['./apps/**/*.{vue,html}', './packages/**/*.vue']
  },
  theme: {
    extend: {
      colors: {
        brand: {
          sky: 'var(--color-brand-sky)',
          night: 'var(--color-brand-night)'
        },
        surface: {
          frost: 'var(--color-surface-frost)',
          soft: 'var(--color-surface-soft)'
        },
        text: {
          primary: 'var(--color-text-primary)',
          secondary: 'var(--color-text-secondary)'
        },
        success: 'var(--color-success)',
        warning: 'var(--color-warning)',
        danger: 'var(--color-danger)'
      },
      boxShadow: {
        frost: '0 32px 80px rgba(24, 56, 108, 0.16)',
        soft: '0 20px 60px rgba(38, 52, 82, 0.10)'
      },
      borderRadius: {
        '4xl': '2rem'
      },
      fontFamily: {
        sans: ['"Noto Sans SC"', '"PingFang SC"', '"Microsoft YaHei"', 'system-ui', 'sans-serif'],
        serif: ['"Noto Serif SC"', '"Songti SC"', '"Source Han Serif SC"', 'serif'],
        mono: ['"JetBrains Mono"', '"SFMono-Regular"', 'monospace']
      },
      backgroundImage: {
        'soft-grid':
          'linear-gradient(rgba(255,255,255,0.24) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.24) 1px, transparent 1px)'
      },
      backgroundSize: {
        grid: '36px 36px'
      }
    }
  },
  plugins: []
};
