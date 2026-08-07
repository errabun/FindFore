import type { ReactNode } from 'react';
import { MantineProvider } from '@mantine/core';
import { MemoryRouter } from 'react-router-dom';
import { render } from '@testing-library/react';
import theme, { cssVariablesResolver } from '../theme';

export function renderWithProviders(ui: ReactNode, { route = '/' } = {}) {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <MantineProvider theme={theme} defaultColorScheme='light' cssVariablesResolver={cssVariablesResolver}>
        {ui}
      </MantineProvider>
    </MemoryRouter>,
  );
}
