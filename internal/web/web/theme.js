(() => {
  'use strict';
  const KEY = 'touchbridge.theme';
  const LIGHT_BG = '#f4f4f1';
  const DARK_BG = '#0f141b';
  const AMOLED_BG = '#000000';
  const query = window.matchMedia ? window.matchMedia('(prefers-color-scheme: light)') : null;

  function stored() {
    try {
      const value = localStorage.getItem(KEY);
      return ['light', 'dark', 'amoled', 'system'].includes(value) ? value : 'system';
    } catch (_) {
      return 'system';
    }
  }

  function resolve(mode) {
    if (mode === 'light' || mode === 'dark' || mode === 'amoled') return mode;
    return query && query.matches ? 'light' : 'dark';
  }

  function apply(mode) {
    const theme = resolve(mode);
    document.documentElement.dataset.theme = theme;
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute('content', theme === 'light' ? LIGHT_BG : theme === 'amoled' ? AMOLED_BG : DARK_BG);
    try {
      if (window.TouchBridgeAndroid && window.TouchBridgeAndroid.setThemeMode) {
        window.TouchBridgeAndroid.setThemeMode(theme === 'light' ? 'light' : 'dark');
      }
    } catch (_) {
      // Browser previews and older hosts simply keep their native bars.
    }
    return theme;
  }

  window.TouchBridgeTheme = { KEY, get: stored, resolve, apply };
  apply(stored());
})();
