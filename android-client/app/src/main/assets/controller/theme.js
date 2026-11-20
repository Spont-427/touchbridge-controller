(() => {
  'use strict';
  const KEY = 'touchbridge.theme';
  const LIGHT_BG = '#f4f4f1';
  const DARK_BG = '#0b0e13';
  const query = window.matchMedia ? window.matchMedia('(prefers-color-scheme: light)') : null;

  function stored() {
    try {
      const value = localStorage.getItem(KEY);
      return value === 'light' || value === 'dark' || value === 'system' ? value : 'system';
    } catch (_) {
      return 'system';
    }
  }

  function resolve(mode) {
    return mode === 'light' || mode === 'dark' ? mode : (query && query.matches ? 'light' : 'dark');
  }

  function apply(mode) {
    const theme = resolve(mode);
    document.documentElement.dataset.theme = theme;
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute('content', theme === 'light' ? LIGHT_BG : DARK_BG);
    try {
      if (window.TouchBridgeAndroid && window.TouchBridgeAndroid.setThemeMode) {
        window.TouchBridgeAndroid.setThemeMode(theme);
      }
    } catch (_) {
      // Browser previews and older hosts simply keep their native bars.
    }
    return theme;
  }

  window.TouchBridgeTheme = { KEY, get: stored, resolve, apply };
  apply(stored());
})();
