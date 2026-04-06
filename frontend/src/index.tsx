import React from 'react';
import { createRoot } from 'react-dom/client';
import '@mantine/core/styles.css';
import '@mantine/dates/styles.css';
import { MantineProvider } from '@mantine/core';
import theme, { cssVariablesResolver } from './theme';
import { getColorScheme } from './adapters/storage/localStorageAdapter';
import './index.css';
import App from './components/App/App';
import * as serviceWorkerRegistration from './serviceWorkerRegistration';

const root = createRoot(document.getElementById('root')!);
root.render(
  <React.StrictMode>
    <MantineProvider
      theme={theme}
      defaultColorScheme={getColorScheme()}
      cssVariablesResolver={cssVariablesResolver}
    >
      <App />
    </MantineProvider>
  </React.StrictMode>
);

serviceWorkerRegistration.register();
