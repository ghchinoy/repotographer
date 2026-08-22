import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import catppuccin from '@catppuccin/starlight';

export default defineConfig({
  site: 'https://ghchinoy.github.io',
  base: '/repotographer',
  integrations: [
    starlight({
      title: 'repotographer',
      description: 'GitHub Repository Cartographer & Concept Graph Explorer',
      logo: {
        src: './src/assets/logo.svg',
        replacesTitle: false,
      },
      social: {
        github: 'https://github.com/ghchinoy/repotographer',
      },
      plugins: [
        catppuccin({
          dark: 'mocha-sapphire',
          light: 'latte-sapphire',
        }),
      ],
      sidebar: [
        {
          label: 'Getting Started',
          autogenerate: { directory: 'getting-started' },
        },
        {
          label: 'User Guides',
          autogenerate: { directory: 'guides' },
        },
        {
          label: 'Reference',
          autogenerate: { directory: 'reference' },
        },
      ],
    }),
  ],
});
