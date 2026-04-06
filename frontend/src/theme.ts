import { createTheme, MantineThemeOverride, MantineColorsTuple, CSSVariablesResolver } from '@mantine/core';

const forest: MantineColorsTuple = [
  '#e9f5ec',
  '#d3ebd9',
  '#a5d6b3',
  '#74c08b',
  '#52a86d',
  '#3d9459',
  '#2E5A2E',
  '#275028',
  '#1f4220',
  '#1a3c1a',
];

const sand: MantineColorsTuple = [
  '#fdfbf7',
  '#f7f3eb',
  '#efe8d8',
  '#e4d9c0',
  '#d6c6a3',
  '#c4a876',
  '#b5975f',
  '#a0844e',
  '#87703f',
  '#6e5b33',
];

const sage: MantineColorsTuple = [
  '#f2f7f3',
  '#e5efe7',
  '#c8dece',
  '#a8cab2',
  '#8ab899',
  '#71a883',
  '#5e9671',
  '#4d7f5e',
  '#3f6a4e',
  '#335740',
];

export const cssVariablesResolver: CSSVariablesResolver = (theme) => ({
  variables: {},
  light: {
    '--ff-bg': theme.colors.sand[0],
    '--ff-header-bg': theme.colors.forest[9],
    '--ff-accent': theme.colors.sand[5],
    '--ff-icon-primary': theme.colors.forest[6],
    '--ff-border': theme.colors.sand[2],
    '--ff-heading': theme.colors.forest[9],
    '--ff-label': theme.colors.forest[8],
    '--ff-link': theme.colors.forest[6],
    '--ff-stat': theme.colors.forest[6],
  },
  dark: {
    '--ff-bg': 'var(--mantine-color-dark-7)',
    '--ff-header-bg': 'var(--mantine-color-dark-8)',
    '--ff-accent': theme.colors.sand[5],
    '--ff-icon-primary': theme.colors.forest[4],
    '--ff-border': 'var(--mantine-color-dark-4)',
    '--ff-heading': theme.colors.forest[3],
    '--ff-label': theme.colors.forest[3],
    '--ff-link': theme.colors.forest[3],
    '--ff-stat': theme.colors.forest[3],
  },
});

const theme: MantineThemeOverride = createTheme({
  primaryColor: 'forest',
  colors: {
    forest,
    sand,
    sage,
  },
  fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
  headings: {
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    fontWeight: '600',
  },
  defaultRadius: 'md',
  shadows: {
    xs: '0 1px 2px rgba(0, 0, 0, 0.04), 0 1px 3px rgba(0, 0, 0, 0.06)',
    sm: '0 1px 3px rgba(0, 0, 0, 0.06), 0 2px 6px rgba(0, 0, 0, 0.08)',
    md: '0 2px 4px rgba(0, 0, 0, 0.06), 0 4px 12px rgba(0, 0, 0, 0.08)',
    lg: '0 4px 8px rgba(0, 0, 0, 0.06), 0 8px 24px rgba(0, 0, 0, 0.10)',
    xl: '0 8px 16px rgba(0, 0, 0, 0.08), 0 16px 48px rgba(0, 0, 0, 0.12)',
  },
  components: {
    Button: {
      defaultProps: {
        radius: 32,
      },
    },
    Card: {
      defaultProps: {
        radius: 16,
      },
    },
    Paper: {
      defaultProps: {
        radius: 16,
        shadow: 'md',
      },
    },
    TextInput: {
      defaultProps: {
        radius: 'md',
      },
    },
    PasswordInput: {
      defaultProps: {
        radius: 'md',
      },
    },
    Select: {
      defaultProps: {
        radius: 'md',
      },
    },
  },
});

export default theme;
