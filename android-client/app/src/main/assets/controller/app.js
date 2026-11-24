(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const core = window.TouchBridgeGamepad;
  const badge = $('connectionBadge');
  const statusText = $('statusText');
  const pinInput = $('pinInput');
  const connectButton = $('connectButton');
  const disconnectButton = $('disconnectButton');
  const hint = $('connectHint');
  const touchpad = $('touchpad');
  const sensitivity = $('sensitivity');
  const sensitivityValue = $('sensitivityValue');
  const latencyBadge = $('latencyBadge');
  const latencyValue = $('latencyValue');
  const gamepadSupport = $('gamepadSupport');
  const gamepadStage = $('gamepadStage');
  const gamepadControls = $('gamepadControls');
  const profileSelect = $('profileSelect');
  const profileDescription = $('profileDescription');
  const editLayoutButton = $('editLayoutButton');
  const layoutEditor = $('layoutEditor');
  const selectedControlName = $('selectedControlName');
  const controlLabel = $('controlLabel');
  const controlAction = $('controlAction');
  const controlSize = $('controlSize');
  const controlSizeValue = $('controlSizeValue');
  const controlOpacity = $('controlOpacity');
  const controlOpacityValue = $('controlOpacityValue');
  const addControlAction = $('addControlAction');
  const addControlButton = $('addControlButton');
  const deleteControlButton = $('deleteControlButton');
  const resetLayoutButton = $('resetLayoutButton');
  const exportLayoutButton = $('exportLayoutButton');
  const importLayoutButton = $('importLayoutButton');
  const importLayoutFile = $('importLayoutFile');
  const editorStatus = $('editorStatus');
  const controlList = $('controlList');
  const editorGridButton = $('editorGridButton');
  const editorSnapButton = $('editorSnapButton');
  const centerModeSelect = $('stickCenterMode');
  const centerRange = $('centerRange');
  const centerRangeValue = $('centerRangeValue');
  const stickDeadzone = $('stickDeadzone');
  const stickDeadzoneValue = $('stickDeadzoneValue');
  const stickCurve = $('stickCurve');
  const homeDeviceSummary = $('homeDeviceSummary');
  const devicesConnectionBadge = $('devicesConnectionBadge');
  const savedServerList = $('savedServerList');
  const serverList = $('serverList');
  const manualAddressInput = $('manualAddressInput');
  const controllerSettingsPanel = $('controllerSettingsPanel');
  const editorAdvancedPanel = $('editorAdvancedPanel');
  const hapticsEnabled = $('hapticsEnabled');
  const triggerModeSelect = $('triggerModeSelect');
  const themeSelect = $('themeSelect');
  const gyroSteeringEnabled = $('gyroSteeringEnabled');
  const gyroInvert = $('gyroInvert');
  const gyroCalibrateButton = $('gyroCalibrateButton');
  const gyroStatus = $('gyroStatus');
  const appVersionValue = $('appVersionValue');
  const aboutServerVersion = $('aboutServerVersion');
  const aboutConnectionState = $('aboutConnectionState');
  const aboutGamepadState = $('aboutGamepadState');
  const profileGrid = $('profileGrid');
  const libraryCount = $('libraryCount');
  const layoutsScreen = $('layoutsScreen');
  const layoutList = $('layoutList');
  const layoutManageCount = $('layoutManageCount');
  const layoutStatus = $('layoutStatus');
  const newLayoutForm = $('newLayoutForm');
  const newLayoutName = $('newLayoutName');
  const newLayoutButton = $('newLayoutButton');
  const cancelNewLayout = $('cancelNewLayout');
  const bottomNav = $('bottomNav');
  const settingsScreen = $('settingsScreen');
  const editorProfileName = $('editorProfileName');
  const controllerProfileName = $('controllerProfileName');
  const controllerProfileSub = $('controllerProfileSub');
  const homeScreen = $('homeScreen');
  const devicesScreen = $('devicesScreen');
  const controllerScreen = $('controllerScreen');
  const editorScreen = $('editorScreen');
  const toolsScreen = $('toolsScreen');
  const appearanceScreen = $('appearanceScreen');
  const controlsScreen = $('controlsScreen');
  const sticksScreen = $('sticksScreen');
  const aboutScreen = $('aboutScreen');

  let socket = null;
  let authenticated = false;
  let gamepadAvailable = false;
  let closeMessage = 'Disconnected. Enter the current PIN to reconnect.';
  let pointer = null;
  let pendingDX = 0;
  let pendingDY = 0;
  let movementFrame = 0;
  let movementRetryTimer = 0;
  let heartbeatTimer = 0;
  let latencyTimer = 0;
  let connectionTimeout = 0;
  let gamepadFrame = 0;
  let gamepadRetryTimer = 0;
  let gamepadDirty = false;
  const latencySamples = [];
  const MAX_PENDING_DELTA = 4000;
  const MAX_WIRE_DELTA = 500;
  const MAX_MOVES_PER_FRAME = 4;
  let scrollRemainder = 0;
  let twoFingerCenterY = null;
  const activePointers = new Map();
  // Track all sources holding an input so a second finger/physical key cannot
  // release a button that another pointer is still holding.
  const mouseButtonHolds = new Map();
  const keyHolds = new Map();

  const gamepadState = { left_x: 0, left_y: 0, right_x: 0, right_y: 0, left_trigger: 0, right_trigger: 0 };
  const gamepadButtons = new Set();
  const activeGamepadControls = new Map();
  const stickGestures = new Map();
  let activeProfile = 'generic';
  let layout = [];
  let editMode = false;
  let selectedControlId = null;
  let layoutDrag = null;
  let editorGridEnabled = true;
  let editorSnapEnabled = true;
  let currentScreen = 'home';
  let editorReturnScreen = 'home';
  let selectedServer = null;
  let connectedServerVersion = '';
  let activeConnectionURL = '';
  let queuedServerSelection = null;
  let rememberedServers = [];
  let discoveredServers = [];
  let pendingDeviceToken = '';
  let lastUsedServerId = '';
  let renamingLayoutId = null;
  let deletingLayoutId = null;
  let wheelGesture = null;
  let gyroListening = false;
  let gyroNeutral = null;
  let gyroLastTilt = 0;
  let gyroFiltered = 0;
  const GYRO_MAX_TILT = 28;
  const WHEEL_MAX_ROTATION = Math.PI * 0.75;

  // Presets keep their hand-tuned art and names. Custom layouts are stored as
  // a small registry plus one layout blob per id under touchbridge.layout.<id>.
  const PRESET_CARDS = Object.freeze({
    generic: { name: 'Xbox-style', subtitle: 'Dual stick', art: 'generic' },
    gta: { name: 'GTA V', subtitle: 'On foot · Driving', art: 'gta' },
    rdr2: { name: 'Red Dead 2', subtitle: 'Dual stick', art: 'rdr' },
    forza: { name: 'Forza Horizon', subtitle: 'Driving · Triggers', art: 'forza' }
  });
  const PROFILE_ART = Object.freeze({
    generic: '<span class="mini-stick"></span><span class="mini-dpad">✣</span><span class="mini-face">A B<br>X Y</span><span class="mini-stick right"></span>',
    gta: '<span class="mini-stick"></span><span class="mini-face">A B<br>X Y</span><span class="mini-stick right"></span>',
    rdr: '<span class="mini-stick"></span><span class="mini-dpad">✣</span><span class="mini-face">A B<br>X Y</span><span class="mini-stick right"></span>',
    forza: '<span class="mini-trigger">LT</span><span class="mini-steering">◎</span><span class="mini-trigger right">RT</span>'
  });
  const PROFILES_KEY = 'touchbridge.profiles';
  const LAYOUT_STATUS_DEFAULT = 'Preset layouts can be duplicated but not deleted.';
  const BUTTON_LABELS = Object.freeze({
    DPAD_UP: 'D-pad up', DPAD_DOWN: 'D-pad down', DPAD_LEFT: 'D-pad left', DPAD_RIGHT: 'D-pad right',
    BACK: 'Back', START: 'Start', GUIDE: 'Guide', LS: 'Stick click L', RS: 'Stick click R'
  });
  const ADD_CLUSTERS = Object.freeze({
    dpad: [
      { button: 'DPAD_UP', label: '▲', dx: 0, dy: -8, size: 43 },
      { button: 'DPAD_LEFT', label: '◀', dx: -7, dy: 0, size: 43 },
      { button: 'DPAD_RIGHT', label: '▶', dx: 7, dy: 0, size: 43 },
      { button: 'DPAD_DOWN', label: '▼', dx: 0, dy: 8, size: 43 }
    ],
    face: [
      { button: 'Y', label: 'Y', dx: 0, dy: -9, size: 46 },
      { button: 'X', label: 'X', dx: -7, dy: 0, size: 46 },
      { button: 'B', label: 'B', dx: 7, dy: 0, size: 46 },
      { button: 'A', label: 'A', dx: 0, dy: 9, size: 46 }
    ],
    shoulders: [
      { button: 'LB', label: 'LB', dx: -8, dy: 0, size: 48 },
      { button: 'RB', label: 'RB', dx: 8, dy: 0, size: 48 }
    ],
    triggers: [
      { button: 'LT', label: 'LT', dx: -8, dy: 0, size: 48 },
      { button: 'RT', label: 'RT', dx: 8, dy: 0, size: 48 }
    ]
  });

  function setStatus(online, message) {
    badge.classList.toggle('online', online);
    badge.classList.toggle('offline', !online);
    statusText.textContent = online ? 'Connected' : 'Not connected';
    connectButton.disabled = online;
    disconnectButton.disabled = !online && !socket;
    pinInput.disabled = online;
    if (hint) hint.textContent = message || (online ? 'Connected. Input is being forwarded to your PC.' : 'Select a computer to connect.');
    if (devicesConnectionBadge) devicesConnectionBadge.textContent = online ? 'CONNECTED' : 'OFFLINE';
    if (homeDeviceSummary) homeDeviceSummary.textContent = online ? (selectedServer?.name || 'Computer connected') : (rememberedServers.length ? `${rememberedServers.length} remembered computer${rememberedServers.length === 1 ? '' : 's'}` : 'No PC connected');
    document.body.classList.toggle('is-connected', online);
    updateAboutDiagnostics();
  }

  function setLatency(value) {
    latencyValue.textContent = value === null ? '— ms' : `${Math.round(value)} ms`;
    latencyBadge.classList.toggle('good', value !== null && value <= 70);
    latencyBadge.classList.toggle('warn', value !== null && value > 70 && value <= 140);
    latencyBadge.classList.toggle('bad', value !== null && value > 140);
    if (value === null) latencyBadge.title = 'Round-trip network latency';
  }

  function recordLatency(sample) {
    latencySamples.push(sample);
    if (latencySamples.length > 32) latencySamples.shift();
    const sorted = [...latencySamples].sort((a, b) => a - b);
    const middle = Math.floor(sorted.length / 2);
    const median = sorted.length % 2 ? sorted[middle] : (sorted[middle - 1] + sorted[middle]) / 2;
    const p95 = sorted[Math.min(sorted.length - 1, Math.ceil(sorted.length * 0.95) - 1)];
    const changes = [];
    for (let i = 1; i < latencySamples.length; i++) changes.push(Math.abs(latencySamples[i] - latencySamples[i - 1]));
    const jitter = changes.length ? changes.reduce((sum, n) => sum + n, 0) / changes.length : 0;
    setLatency(median);
    latencyBadge.title = `RTT median ${median.toFixed(1)} ms · p95 ${p95.toFixed(1)} ms · jitter ${jitter.toFixed(1)} ms`;
  }

  function send(payload) {
    if (!socket || socket.readyState !== WebSocket.OPEN || !authenticated) return false;
    // Keep the send queue bounded. Mouse deltas are accumulated; gamepad
    // snapshots are coalesced to the newest state by flushGamepadState().
    if (socket.bufferedAmount > 32768 && (payload.type === 'mouse_move' || payload.type === 'gamepad_state')) return false;
    socket.send(JSON.stringify(payload));
    return true;
  }

  function holdInput(registry, name, token, type) {
    let owners = registry.get(name);
    if (!owners) {
      owners = new Set();
      registry.set(name, owners);
    }
    if (owners.has(token)) return;
    const wasReleased = owners.size === 0;
    owners.add(token);
    if (wasReleased) send({ type, [type === 'key' ? 'key' : 'button']: name, down: true });
  }

  function releaseInput(registry, name, token, type) {
    const owners = registry.get(name);
    if (!owners || !owners.has(token)) return;
    owners.delete(token);
    if (owners.size === 0) {
      registry.delete(name);
      send({ type, [type === 'key' ? 'key' : 'button']: name, down: false });
    }
  }

  function scheduleMovementFlush(delay = 0) {
    if (movementFrame || movementRetryTimer || (!Math.trunc(pendingDX) && !Math.trunc(pendingDY))) return;
    if (delay > 0) {
      movementRetryTimer = window.setTimeout(() => {
        movementRetryTimer = 0;
        scheduleMovementFlush();
      }, delay);
    } else {
      movementFrame = requestAnimationFrame(flushMovement);
    }
  }

  function flushMovement() {
    movementFrame = 0;
    if (!socket || socket.readyState !== WebSocket.OPEN || !authenticated) {
      pendingDX = pendingDY = 0;
      return;
    }
    if (socket.bufferedAmount > 32768) {
      scheduleMovementFlush(24);
      return;
    }
    let sent = 0;
    while (sent < MAX_MOVES_PER_FRAME) {
      const dx = Math.trunc(pendingDX);
      const dy = Math.trunc(pendingDY);
      if (!dx && !dy) break;
      const chunkDX = Math.max(-MAX_WIRE_DELTA, Math.min(MAX_WIRE_DELTA, dx));
      const chunkDY = Math.max(-MAX_WIRE_DELTA, Math.min(MAX_WIRE_DELTA, dy));
      if (!send({ type: 'mouse_move', dx: chunkDX, dy: chunkDY })) break;
      pendingDX -= chunkDX;
      pendingDY -= chunkDY;
      sent += 1;
    }
    if (Math.trunc(pendingDX) || Math.trunc(pendingDY)) scheduleMovementFlush(socket.bufferedAmount > 32768 ? 24 : 0);
  }

  function queueMovement(dx, dy) {
    pendingDX = Math.max(-MAX_PENDING_DELTA, Math.min(MAX_PENDING_DELTA, pendingDX + dx));
    pendingDY = Math.max(-MAX_PENDING_DELTA, Math.min(MAX_PENDING_DELTA, pendingDY + dy));
    scheduleMovementFlush();
  }

  function setNativeOrientation(mode) {
    if (window.TouchBridgeAndroid?.setScreenMode) window.TouchBridgeAndroid.setScreenMode(mode);
    document.body.classList.toggle('tb-landscape', mode === 'landscape');
  }

  function applyThemePreference(mode) {
    const value = ['light', 'dark', 'amoled'].includes(mode) ? mode : 'system';
    safeStorageSet('touchbridge.theme', value);
    if (themeSelect) themeSelect.value = value;
    const resolved = window.TouchBridgeTheme
      ? window.TouchBridgeTheme.apply(value)
      : (value === 'light' ? 'light' : 'dark');
    document.documentElement.dataset.theme = resolved;
  }

  function updateAboutDiagnostics() {
    if (aboutConnectionState) aboutConnectionState.textContent = authenticated ? 'Connected' : 'Disconnected';
    if (aboutServerVersion) aboutServerVersion.textContent = authenticated ? (connectedServerVersion || 'Connected') : '—';
    if (aboutGamepadState) aboutGamepadState.textContent = gamepadAvailable ? 'Available' : (authenticated ? 'Unavailable' : 'Unknown');
  }

  function mountGamepadStage(mountId) {
    const mount = $(mountId);
    if (mount && gamepadStage.parentElement !== mount) mount.append(gamepadStage);
  }

  function updateProfileLabels() {
    const meta = core?.PROFILE_META[activeProfile];
    const label = displayName(activeProfile);
    if (controllerProfileName) controllerProfileName.textContent = label;
    if (editorProfileName) editorProfileName.textContent = `${label} layout`;
    if (controllerProfileSub) controllerProfileSub.textContent = 'GAME CONTROLLER';
    if (profileDescription) profileDescription.textContent = meta?.description || 'Custom layout saved on this phone.';
  }

  function setProfile(profileId) {
    if (!core || !profileExists(profileId)) profileId = 'generic';
    if (stickGestures.size || activeGamepadControls.size) clearGamepadState(true, false);
    activeProfile = profileId;
    profileSelect.value = profileId;
    layout = loadLayout(activeProfile);
    selectedControlId = null;
    safeStorageSet('touchbridge.lastProfile', activeProfile);
    updateProfileLabels();
    renderGamepadLayout();
  }

  function setScreen(screen) {
    currentScreen = screen;
    homeScreen.hidden = screen !== 'home';
    devicesScreen.hidden = screen !== 'devices';
    controllerScreen.hidden = screen !== 'controller';
    editorScreen.hidden = screen !== 'editor';
    settingsScreen.hidden = screen !== 'settings';
    appearanceScreen.hidden = screen !== 'appearance';
    controlsScreen.hidden = screen !== 'controls';
    sticksScreen.hidden = screen !== 'sticks';
    aboutScreen.hidden = screen !== 'about';
    layoutsScreen.hidden = screen !== 'layouts';
    toolsScreen.hidden = screen !== 'tools';
    if (bottomNav) {
      bottomNav.hidden = screen === 'controller' || screen === 'editor';
      const settingsChildren = ['tools', 'layouts', 'appearance', 'controls', 'sticks', 'about'];
      const active = settingsChildren.includes(screen) ? 'settings' : screen;
      bottomNav.querySelectorAll('[data-nav-screen]').forEach(button => {
        const selected = button.dataset.navScreen === active;
        button.classList.toggle('active', selected);
        if (selected) button.setAttribute('aria-current', 'page');
        else button.removeAttribute('aria-current');
      });
    }
    if (screen === 'controller') {
      if (editMode) setEditorMode(false);
      mountGamepadStage('controllerStageMount');
      if (controllerSettingsPanel) controllerSettingsPanel.hidden = true;
      setNativeOrientation('landscape');
    } else if (screen === 'editor') {
      setEditorMode(true);
      mountGamepadStage('editorStageMount');
      if (editorAdvancedPanel) editorAdvancedPanel.hidden = true;
      if (layoutEditor) layoutEditor.hidden = true;
      setNativeOrientation('landscape');
    } else {
      if (editMode) setEditorMode(false);
      mountGamepadStage('controllerStageMount');
      if (controllerSettingsPanel) controllerSettingsPanel.hidden = true;
      if (editorAdvancedPanel) editorAdvancedPanel.hidden = true;
      setNativeOrientation('portrait');
    }
    if (screen === 'devices') renderDeviceLists();
    if (screen === 'layouts') renderLayoutManager();
    if (screen === 'about') updateAboutDiagnostics();
    if (screen === 'controller' || screen === 'editor') {
      updateProfileLabels();
      if (profileSelect) profileSelect.value = activeProfile;
    }
  }

  function openEditor() {
    editorReturnScreen = currentScreen === 'controller' ? 'controller' : currentScreen === 'layouts' ? 'layouts' : 'home';
    setScreen('editor');
  }

  function openProfile(profileId, mode) {
    if (profileId) setProfile(profileId);
    if (mode === 'edit') openEditor();
    else setScreen('controller');
  }

  function scanDevices() {
    if (window.TouchBridgeAndroid?.scanServers) {
      if (serverList) serverList.innerHTML = '<p class="empty-state">Scanning the local network…</p>';
      window.TouchBridgeAndroid.scanServers();
      return;
    }
    const isServerOrigin = location.hostname && location.hostname !== 'touchbridge.appassets.local' && location.protocol.startsWith('http');
    if (isServerOrigin) {
      discoveredServers = [{ name: 'This computer', url: location.origin, serverId: '', address: location.hostname, port: Number(location.port || (location.protocol === 'https:' ? 443 : 80)) }];
    }
    renderDeviceLists();
  }

  function onServersFound(servers) {
    discoveredServers = Array.isArray(servers) ? servers.filter(item => item && normaliseServerURL(item.url)) : [];
    // A server can receive a new DHCP address. Its persistent server ID lets
    // this phone update the address while retaining its locally stored token.
    for (const found of discoveredServers) {
      const saved = rememberedServers.find(item => item.serverId && found.serverId && item.serverId === found.serverId);
      if (saved && saved.url !== found.url) {
        saved.url = normaliseServerURL(found.url);
        if (window.TouchBridgeAndroid?.saveTrustedServer) {
          window.TouchBridgeAndroid.saveTrustedServer(saved.url, saved.name, saved.serverId, saved.token || '');
        }
      }
    }
    rememberedServers = getRememberedServers();
    renderDeviceLists();
  }

  window.TouchBridge = {
    onServersFound,
    androidBack() {
      if (currentScreen === 'editor') {
        if (editorAdvancedPanel && !editorAdvancedPanel.hidden) { editorAdvancedPanel.hidden = true; return true; }
        setScreen(editorReturnScreen);
        return true;
      }
      if (currentScreen === 'controller') { setScreen('home'); return true; }
      if (['layouts', 'appearance', 'controls', 'sticks', 'about', 'tools'].includes(currentScreen)) { setScreen('settings'); return true; }
      if (currentScreen === 'devices' || currentScreen === 'settings') { setScreen('home'); return true; }
      if (currentScreen === 'home') return false;
      return false;
    }
  };

  $('homeDevicesButton').addEventListener('click', () => { setScreen('devices'); scanDevices(); });
  $('toolsBackButton').addEventListener('click', () => setScreen('settings'));
  $('settingsAppearanceButton').addEventListener('click', () => setScreen('appearance'));
  $('settingsControlsButton').addEventListener('click', () => setScreen('controls'));
  $('settingsSticksButton').addEventListener('click', () => setScreen('sticks'));
  $('settingsAboutButton').addEventListener('click', () => setScreen('about'));
  $('appearanceBackButton').addEventListener('click', () => setScreen('settings'));
  $('controlsBackButton').addEventListener('click', () => setScreen('settings'));
  $('sticksBackButton').addEventListener('click', () => setScreen('settings'));
  $('aboutBackButton').addEventListener('click', () => setScreen('settings'));
  $('scanButton').addEventListener('click', scanDevices);
  $('settingsToolsButton').addEventListener('click', () => setScreen('tools'));
  $('settingsLayoutsButton').addEventListener('click', () => {
    renamingLayoutId = null;
    deletingLayoutId = null;
    newLayoutForm.hidden = true;
    newLayoutButton.disabled = false;
    setScreen('layouts');
    setLayoutStatus(LAYOUT_STATUS_DEFAULT);
  });
  $('layoutsBackButton').addEventListener('click', () => setScreen('settings'));
  newLayoutButton.addEventListener('click', () => {
    newLayoutForm.hidden = false;
    newLayoutButton.disabled = true;
    newLayoutName.value = '';
    newLayoutName.focus();
    setLayoutStatus('Name the new layout, then choose Create.');
  });
  cancelNewLayout.addEventListener('click', () => {
    newLayoutForm.hidden = true;
    newLayoutButton.disabled = false;
    newLayoutName.value = '';
    setLayoutStatus(LAYOUT_STATUS_DEFAULT);
  });
  newLayoutForm.addEventListener('submit', event => {
    event.preventDefault();
    createLayout(newLayoutName.value);
  });
  layoutList.addEventListener('click', event => {
    const button = event.target.closest('button[data-layout-action]');
    if (!button) return;
    const row = button.closest('[data-layout-id]');
    const id = row?.dataset.layoutId;
    if (!id) return;
    const action = button.dataset.layoutAction;
    if (action === 'edit') { openProfile(id, 'edit'); return; }
    if (action === 'duplicate') { duplicateLayout(id); return; }
    if (action === 'rename') { renamingLayoutId = id; deletingLayoutId = null; renderLayoutManager(); setLayoutStatus('Enter a new name, then choose Save.'); return; }
    if (action === 'delete') { deletingLayoutId = id; renamingLayoutId = null; renderLayoutManager(); return; }
    if (action === 'cancel-rename' || action === 'cancel-delete') { renamingLayoutId = null; deletingLayoutId = null; renderLayoutManager(); setLayoutStatus(LAYOUT_STATUS_DEFAULT); return; }
    if (action === 'save-rename') { renameLayout(id, row.querySelector('.layout-rename-input')?.value || ''); return; }
    if (action === 'confirm-delete') { deleteLayout(id); }
  });
  layoutList.addEventListener('keydown', event => {
    if (event.key !== 'Enter' || !event.target.matches('.layout-rename-input')) return;
    event.preventDefault();
    const row = event.target.closest('[data-layout-id]');
    if (row) renameLayout(row.dataset.layoutId, event.target.value);
  });
  $('connectionBadge').addEventListener('click', () => { setScreen('devices'); scanDevices(); });
  $('editorDoneButton').addEventListener('click', () => { setEditorMode(false); setScreen(editorReturnScreen); });
  $('closeEditorAdvanced').addEventListener('click', () => {
    editorAdvancedPanel.hidden = true;
    $('editorMoreButton').setAttribute('aria-expanded', 'false');
  });
  $('editorMoreButton').addEventListener('click', () => {
    editorAdvancedPanel.hidden = !editorAdvancedPanel.hidden;
    $('editorMoreButton').setAttribute('aria-expanded', String(!editorAdvancedPanel.hidden));
  });
  bottomNav?.querySelectorAll('[data-nav-screen]').forEach(button => {
    button.addEventListener('click', () => {
      const target = button.dataset.navScreen;
      setScreen(target);
      if (target === 'devices') scanDevices();
    });
  });
  $('manualAddressButton').addEventListener('click', () => {
    const url = normaliseServerURL(manualAddressInput.value.trim());
    if (!url) { hint.textContent = 'Enter a valid HTTP or HTTPS computer address.'; return; }
    selectServer({ name: new URL(url).hostname, url, serverId: '', token: '' });
  });
  profileGrid?.addEventListener('click', event => {
    const play = event.target.closest('[data-play-profile]');
    if (play) { openProfile(play.dataset.playProfile, 'play'); return; }
    const edit = event.target.closest('[data-edit-profile]');
    if (edit) { openProfile(edit.dataset.editProfile, 'edit'); return; }
    const card = event.target.closest('[data-profile-card]');
    if (card) openProfile(card.dataset.profileCard, 'play');
  });

  function startLatencyProbe() {
    if (latencyTimer) clearInterval(latencyTimer);
    latencySamples.length = 0;
    setLatency(null);
    latencyTimer = setInterval(() => {
      if (authenticated) send({ type: 'latency_ping', timestamp: performance.now() });
    }, 2500);
  }

  function normaliseServerURL(value) {
    try {
      const url = new URL(value);
      if (!['http:', 'https:'].includes(url.protocol) || !url.hostname || url.username || url.password) return '';
      url.pathname = '/'; url.search = ''; url.hash = '';
      return url.origin;
    } catch { return ''; }
  }

  function websocketURL(serverURL) {
    const url = new URL('/ws', serverURL);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    return url.toString();
  }

  function getRememberedServers() {
    if (window.TouchBridgeAndroid?.getSavedServersJson) {
      try {
        const parsed = JSON.parse(window.TouchBridgeAndroid.getSavedServersJson());
        return Array.isArray(parsed) ? parsed : [];
      } catch { return []; }
    }
    try {
      const parsed = JSON.parse(localStorage.getItem('touchbridge.savedServers') || '[]');
      return Array.isArray(parsed) ? parsed : [];
    } catch { return []; }
  }

  function persistRememberedServer(server, token) {
    const record = {
      url: normaliseServerURL(server.url),
      name: String(server.name || 'TouchBridge computer').slice(0, 64),
      serverId: String(server.serverId || ''),
      token: String(token || ''),
      lastUsed: Date.now()
    };
    if (!record.url) return;
    if (window.TouchBridgeAndroid?.saveTrustedServer) {
      window.TouchBridgeAndroid.saveTrustedServer(record.url, record.name, record.serverId, record.token);
    } else {
      const next = rememberedServers.filter(item => !(record.serverId && item.serverId === record.serverId) && item.url !== record.url);
      next.unshift(record);
      try { localStorage.setItem('touchbridge.savedServers', JSON.stringify(next.slice(0, 16))); } catch { /* storage quota or private mode */ }
    }
    rememberedServers = getRememberedServers();
    renderDeviceLists();
  }

  function forgetRememberedServer(server) {
    if (!server) return;
    if (window.TouchBridgeAndroid?.forgetTrustedServer) {
      window.TouchBridgeAndroid.forgetTrustedServer(server.serverId || server.url);
    } else {
      rememberedServers = rememberedServers.filter(item => item.url !== server.url && (!server.serverId || item.serverId !== server.serverId));
      try { localStorage.setItem('touchbridge.savedServers', JSON.stringify(rememberedServers)); } catch { /* ignore */ }
    }
    rememberedServers = getRememberedServers();
    renderDeviceLists();
  }

  function renderDeviceCard(server, remembered) {
    const card = document.createElement('article');
    card.className = 'device-row';
    const icon = document.createElement('span'); icon.className = 'device-icon'; icon.textContent = '▣';
    const meta = document.createElement('span'); meta.className = 'device-meta';
    const name = document.createElement('strong'); name.textContent = server.name || 'TouchBridge computer';
    const address = document.createElement('small'); address.textContent = String(server.url || '').replace(/^https?:\/\//, '');
    meta.append(name, address);
    const connect = document.createElement('button'); connect.type = 'button'; connect.textContent = remembered && server.token ? 'Connect' : (remembered ? 'Pair again' : 'Pair');
    connect.addEventListener('click', () => selectServer(server));
    card.append(icon, meta, connect);
    if (remembered) {
      const forget = document.createElement('button'); forget.type = 'button'; forget.className = 'forget-device'; forget.textContent = 'Forget'; forget.title = 'Forget this phone’s saved connection';
      forget.addEventListener('click', () => forgetRememberedServer(server));
      card.append(forget);
    }
    return card;
  }

  function renderDeviceLists() {
    rememberedServers = getRememberedServers();
    if (savedServerList) {
      savedServerList.replaceChildren();
      if (!rememberedServers.length) {
        const empty = document.createElement('p'); empty.className = 'empty-state'; empty.textContent = 'No remembered computers yet. Pair a computer once and it will appear here.';
        savedServerList.append(empty);
      } else rememberedServers.forEach(server => savedServerList.append(renderDeviceCard(server, true)));
    }
    if (serverList) {
      serverList.replaceChildren();
      const rememberedIDs = new Set(rememberedServers.map(item => item.serverId).filter(Boolean));
      const fresh = discoveredServers.filter(item => !rememberedIDs.has(item.serverId));
      if (!fresh.length) {
        const empty = document.createElement('p'); empty.className = 'empty-state'; empty.textContent = 'No new computers found. Try scanning again or enter an address manually.';
        serverList.append(empty);
      } else fresh.forEach(server => serverList.append(renderDeviceCard(server, false)));
    }
    if (homeDeviceSummary && currentScreen === 'home') {
      homeDeviceSummary.textContent = rememberedServers.length ? `${rememberedServers.length} remembered computer${rememberedServers.length === 1 ? '' : 's'}` : 'No PC connected';
    }
  }

  function selectServer(server) {
    const url = normaliseServerURL(server.url || `http://${server.address}:${server.port || 8765}`);
    if (!url) { hint.textContent = 'That computer address is invalid.'; return; }
    const target = { ...server, url };

    if (socket && socket.readyState !== WebSocket.CLOSED) {
      if (activeConnectionURL === url) {
        if (authenticated) setScreen('controller');
        else setScreen('devices');
        return;
      }
      // Switch cleanly: release held desktop inputs, close the old socket, and
      // only begin authenticating to the new PC after the old connection exits.
      queuedServerSelection = target;
      hint.textContent = 'Switching computers…';
      send({ type: 'release_all' });
      socket.close();
      return;
    }

    selectedServer = target;
    pendingDeviceToken = target.token || '';
    pinInput.value = '';
    $('pairTitle').textContent = `Pair with ${target.name || 'computer'}`;
    $('pairDescription').textContent = pendingDeviceToken
      ? 'This phone is already trusted by this computer. Connecting without another PIN…'
      : 'Enter the six-digit PIN shown in the computer’s TouchBridge terminal. This is normally required only once.';
    setScreen('devices');
    if (pendingDeviceToken) connect();
    else { hint.textContent = 'Enter the PIN shown on your computer. It will remember this phone after pairing.'; pinInput.focus(); }
  }

  function connect() {
    if (socket && socket.readyState !== WebSocket.CLOSED) return;
    const currentOriginIsServer = location.hostname && location.hostname !== 'touchbridge.appassets.local' && location.protocol.startsWith('http');
    const targetURL = selectedServer?.url || (currentOriginIsServer ? location.origin : '');
    if (!selectedServer && currentOriginIsServer) selectedServer = { name: 'This computer', url: targetURL, serverId: '', token: '' };
    if (!targetURL) {
      hint.textContent = 'Choose a nearby or remembered computer first.';
      return;
    }

    const connectionServer = selectedServer ? { ...selectedServer } : null;
    const pin = pinInput.value.trim();
    const token = pendingDeviceToken || connectionServer?.token || (connectionServer?.serverId ? rememberedServers.find(item => item.serverId === connectionServer.serverId)?.token : '') || rememberedServers.find(item => item.url === targetURL)?.token || '';
    if (!token && !/^\d{6}$/.test(pin)) {
      hint.textContent = 'Enter the six-digit PIN shown in the computer’s TouchBridge terminal.';
      pinInput.focus();
      return;
    }
    const authMessage = token
      ? { type: 'auth_device', device_token: token }
      : { type: 'auth', pin, remember_device: true, device_name: 'TouchBridge Android' };
    closeMessage = 'Disconnected. Select a remembered computer to reconnect.';
    let ws;
    try {
      ws = new WebSocket(websocketURL(targetURL));
    } catch (error) {
      const message = error && error.message ? error.message : 'Invalid server address.';
      hint.textContent = `Could not open connection: ${message}`;
      setStatus(false, hint.textContent);
      return;
    }
    if (connectionTimeout) clearTimeout(connectionTimeout);
    connectionTimeout = window.setTimeout(() => {
      if (socket !== ws || authenticated) return;
      closeMessage = 'Connection timed out. Check the PC address, Wi-Fi and firewall.';
      hint.textContent = closeMessage;
      ws.close();
    }, 10000);
    socket = ws;
    activeConnectionURL = targetURL;
    setStatus(false, token ? 'Reconnecting to your trusted computer…' : 'Pairing with your computer…');

    ws.addEventListener('open', () => {
      if (socket === ws && !queuedServerSelection) ws.send(JSON.stringify(authMessage));
    });
    ws.addEventListener('message', (event) => {
      // Ignore late messages from a socket being replaced by another PC.
      if (socket !== ws || queuedServerSelection) return;
      let data;
      try { data = JSON.parse(event.data); } catch { return; }
      if (data.type === 'auth_ok') {
        if (connectionTimeout) clearTimeout(connectionTimeout);
        connectionTimeout = 0;
        authenticated = true;
        connectedServerVersion = String(data.server_version || '');
        pendingDX = pendingDY = 0;
        if (data.device_token && connectionServer) {
          connectionServer.serverId = data.server_id || connectionServer.serverId || '';
          connectionServer.token = data.device_token;
          selectedServer = connectionServer;
          persistRememberedServer(connectionServer, data.device_token);
        } else if (connectionServer && data.server_id && !connectionServer.serverId) {
          connectionServer.serverId = data.server_id;
          selectedServer = connectionServer;
          persistRememberedServer(connectionServer, connectionServer.token || '');
        }
        pendingDeviceToken = connectionServer?.token || '';
        if (heartbeatTimer) clearInterval(heartbeatTimer);
        heartbeatTimer = setInterval(() => send({ type: 'keepalive' }), 15000);
        startLatencyProbe();
        closeMessage = 'Disconnected. Select a remembered computer to reconnect.';
        setStatus(true, 'Connected. Gamepad input is ready.');
        pinInput.value = '';
        if (currentScreen === 'devices') setScreen('controller');
      } else if (data.type === 'capabilities') {
        setGamepadAvailability(Boolean(data.gamepad), data.gamepad_status || '');
      } else if (data.type === 'latency_pong') {
        const started = Number(data.timestamp);
        if (Number.isFinite(started) && started >= 0) recordLatency(Math.max(0, performance.now() - started));
      } else if (data.type === 'auth_error') {
        authenticated = false;
        if (token) {
          if (connectionServer) forgetRememberedServer(connectionServer);
          if (selectedServer?.url === connectionServer?.url) selectedServer.token = '';
          if (connectionServer) connectionServer.token = '';
          pendingDeviceToken = '';
          $('pairDescription').textContent = 'The saved connection is no longer trusted. Enter the current six-digit PIN to pair this phone again.';
          closeMessage = 'The saved trust token is no longer valid. Pair this computer again with its six-digit PIN.';
        } else closeMessage = 'PIN rejected or expired. Check the computer terminal for the current PIN.';
        hint.textContent = closeMessage;
        ws.close();
      } else if (data.type === 'rate_limited') {
        authenticated = false;
        closeMessage = 'Too many incorrect PIN attempts. Wait one minute before trying again.';
        hint.textContent = closeMessage;
        ws.close();
      } else if (data.type === 'controller_busy') {
        authenticated = false;
        closeMessage = 'Another phone is already controlling this PC. Disconnect it first.';
        hint.textContent = closeMessage;
        ws.close();
      }
    });
    ws.addEventListener('close', () => {
      if (socket !== ws) return;
      if (connectionTimeout) clearTimeout(connectionTimeout);
      connectionTimeout = 0;
      authenticated = false;
      gamepadAvailable = false;
      connectedServerVersion = '';
      setGamepadAvailability(false, 'Connect a computer to send controller input.');
      if (heartbeatTimer) clearInterval(heartbeatTimer);
      if (latencyTimer) clearInterval(latencyTimer);
      heartbeatTimer = latencyTimer = 0;
      latencySamples.length = 0;
      setLatency(null);
      socket = null;
      activeConnectionURL = '';
      pendingDX = pendingDY = 0;
      if (movementFrame) cancelAnimationFrame(movementFrame);
      if (movementRetryTimer) clearTimeout(movementRetryTimer);
      movementFrame = movementRetryTimer = 0;
      if (gamepadFrame) cancelAnimationFrame(gamepadFrame);
      if (gamepadRetryTimer) clearTimeout(gamepadRetryTimer);
      gamepadFrame = gamepadRetryTimer = 0;
      gamepadDirty = false;
      clearGamepadState(false, false);
      releaseLocalState();
      setStatus(false, closeMessage);

      if (queuedServerSelection) {
        const next = queuedServerSelection;
        queuedServerSelection = null;
        selectedServer = next;
        pendingDeviceToken = next.token || '';
        pinInput.value = '';
        $('pairTitle').textContent = `Pair with ${next.name || 'computer'}`;
        $('pairDescription').textContent = pendingDeviceToken
          ? 'This phone is already trusted by this computer. Connecting without another PIN…'
          : 'Enter the six-digit PIN shown in the computer’s TouchBridge terminal. This is normally required only once.';
        setScreen('devices');
        if (pendingDeviceToken) connect();
        else {
          hint.textContent = 'Enter the PIN shown on your computer. It will remember this phone after pairing.';
          pinInput.focus();
        }
      }
    });
    ws.addEventListener('error', () => {
      if (socket !== ws) return;
      closeMessage = 'Connection failed. Check Wi-Fi and the PC firewall.';
      hint.textContent = closeMessage;
    });
  }

  function disconnect() {
    if (socket) {
      send({ type: 'release_all' });
      socket.close();
    }
  }

  function releaseLocalState() {
    mouseButtonHolds.clear();
    keyHolds.clear();
    activePointers.clear();
    pointer = null;
    twoFingerCenterY = null;
    scrollRemainder = 0;
    stickGestures.clear();
    activeGamepadControls.clear();
    layoutDrag = null;
    document.querySelectorAll('.pressed').forEach(el => el.classList.remove('pressed'));
    touchpad.classList.remove('active');
    clearGamepadState(false, false);
    renderGamepadLayout();
  }

  connectButton.addEventListener('click', connect);
  disconnectButton.addEventListener('click', disconnect);
  pinInput.addEventListener('input', () => { pinInput.value = pinInput.value.replace(/\D/g, '').slice(0, 6); });
  pinInput.addEventListener('keydown', e => { if (e.key === 'Enter') connect(); });
  sensitivity.addEventListener('input', () => { sensitivityValue.value = `${Number(sensitivity.value).toFixed(1)}×`; safeStorageSet('touchbridge.mouse.sensitivity', sensitivity.value); });

  document.querySelectorAll('.tools-screen [data-view]').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.tools-screen [data-view]').forEach(t => t.classList.toggle('active', t === tab));
      document.querySelectorAll('.tools-screen .mode-view').forEach(view => view.classList.toggle('active-view', view.id === tab.dataset.view));
    });
  });

  // Mouse / touchpad input.
  touchpad.addEventListener('contextmenu', e => e.preventDefault());
  touchpad.addEventListener('pointerdown', e => {
    if (e.button !== 0) return;
    e.preventDefault();
    touchpad.setPointerCapture(e.pointerId);
    activePointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (activePointers.size === 1) {
      pointer = { id: e.pointerId, x: e.clientX, y: e.clientY, startX: e.clientX, startY: e.clientY, started: performance.now(), moved: false };
      touchpad.classList.add('active');
    } else if (activePointers.size === 2) {
      const values = [...activePointers.values()];
      twoFingerCenterY = (values[0].y + values[1].y) / 2;
      if (pointer) pointer.moved = true;
    }
  });
  touchpad.addEventListener('pointermove', e => {
    if (!activePointers.has(e.pointerId)) return;
    const samples = typeof e.getCoalescedEvents === 'function' ? e.getCoalescedEvents() : [e];
    if (activePointers.size >= 2) {
      const current = samples.length ? samples[samples.length - 1] : e;
      activePointers.set(e.pointerId, { x: current.clientX, y: current.clientY });
      const values = [...activePointers.values()];
      const centerY = values.reduce((sum, point) => sum + point.y, 0) / values.length;
      if (twoFingerCenterY !== null) scrollRemainder += (centerY - twoFingerCenterY) / 24;
      twoFingerCenterY = centerY;
      const amount = Math.trunc(scrollRemainder);
      if (amount) { scrollRemainder -= amount; send({ type: 'scroll', amount: Math.max(-20, Math.min(20, amount)) }); }
      return;
    }
    if (!pointer || pointer.id !== e.pointerId) return;
    const scale = Number(sensitivity.value);
    let dx = 0, dy = 0, x = pointer.x, y = pointer.y;
    for (const sample of samples) {
      dx += (sample.clientX - x) * scale;
      dy += (sample.clientY - y) * scale;
      x = sample.clientX;
      y = sample.clientY;
    }
    pointer.x = x;
    pointer.y = y;
    if (Math.hypot(x - pointer.startX, y - pointer.startY) > 7) pointer.moved = true;
    activePointers.set(e.pointerId, { x, y });
    if (dx || dy) queueMovement(dx, dy);
  });
  function endPointer(e) {
    const wasSinglePointer = activePointers.size === 1 && pointer && pointer.id === e.pointerId;
    if (e.type === 'pointerup' && wasSinglePointer && !pointer.moved && performance.now() - pointer.started < 260) {
      send({ type: 'mouse_button', button: 'left', down: true });
      send({ type: 'mouse_button', button: 'left', down: false });
    }
    activePointers.delete(e.pointerId);
    if (activePointers.size === 0) {
      pointer = null;
      twoFingerCenterY = null;
      touchpad.classList.remove('active');
    } else if (activePointers.size === 1) {
      const [id, p] = activePointers.entries().next().value;
      pointer = { id, x: p.x, y: p.y, startX: p.x, startY: p.y, started: 0, moved: true };
      twoFingerCenterY = null;
    } else {
      const values = [...activePointers.values()];
      twoFingerCenterY = values.reduce((sum, point) => sum + point.y, 0) / values.length;
    }
  }
  touchpad.addEventListener('pointerup', endPointer);
  touchpad.addEventListener('pointercancel', endPointer);
  touchpad.addEventListener('lostpointercapture', endPointer);
  touchpad.addEventListener('wheel', e => {
    e.preventDefault();
    scrollRemainder += e.deltaY / 80;
    const notches = Math.trunc(scrollRemainder);
    if (notches) { scrollRemainder -= notches; send({ type: 'scroll', amount: Math.max(-20, Math.min(20, notches)) }); }
  }, { passive: false });

  document.querySelectorAll('[data-button]').forEach(button => {
    const name = button.dataset.button;
    const down = e => {
      e.preventDefault();
      if (e.button !== 0 && e.pointerType === 'mouse') return;
      button.setPointerCapture(e.pointerId);
      holdInput(mouseButtonHolds, name, `pointer:${e.pointerId}`, 'mouse_button');
      button.classList.add('pressed');
    };
    const up = e => {
      e.preventDefault();
      releaseInput(mouseButtonHolds, name, `pointer:${e.pointerId}`, 'mouse_button');
      if (!mouseButtonHolds.has(name)) button.classList.remove('pressed');
    };
    button.addEventListener('pointerdown', down);
    button.addEventListener('pointerup', up);
    button.addEventListener('pointercancel', up);
    button.addEventListener('lostpointercapture', up);
  });

  document.querySelectorAll('[data-key]').forEach(button => {
    const key = button.dataset.key;
    const down = e => {
      e.preventDefault();
      if (e.button !== 0 && e.pointerType === 'mouse') return;
      button.setPointerCapture(e.pointerId);
      holdInput(keyHolds, key, `pointer:${e.pointerId}`, 'key');
      button.classList.add('pressed');
    };
    const up = e => {
      e.preventDefault();
      releaseInput(keyHolds, key, `pointer:${e.pointerId}`, 'key');
      if (!keyHolds.has(key)) button.classList.remove('pressed');
    };
    button.addEventListener('pointerdown', down);
    button.addEventListener('pointerup', up);
    button.addEventListener('pointercancel', up);
    button.addEventListener('lostpointercapture', up);
  });

  $('scrollUp').addEventListener('click', () => send({ type: 'scroll', amount: -3 }));
  $('scrollDown').addEventListener('click', () => send({ type: 'scroll', amount: 3 }));

  const keyMap = { Escape: 'ESC', Enter: 'ENTER', ' ': 'SPACE', Tab: 'TAB', Backspace: 'BACKSPACE', Delete: 'DELETE', Shift: 'SHIFT', Control: 'CTRL', Alt: 'ALT', ArrowUp: 'UP', ArrowDown: 'DOWN', ArrowLeft: 'LEFT', ArrowRight: 'RIGHT', Home: 'HOME', End: 'END', PageUp: 'PAGEUP', PageDown: 'PAGEDOWN' };
  function mapKey(e) {
    const raw = e.key.length === 1 ? e.key.toUpperCase() : e.key;
    return keyMap[e.key] || (/^[A-Z0-9]$/.test(raw) ? raw : null) || (/^F([1-9]|1[0-2])$/.test(e.key) ? e.key : null);
  }
  window.addEventListener('keydown', e => {
    if (core?.isInteractiveTarget(e.target)) return;
    const key = mapKey(e);
    if (!key || e.repeat) return;
    e.preventDefault();
    holdInput(keyHolds, key, `hardware:${e.code || e.key}`, 'key');
  });
  window.addEventListener('keyup', e => {
    const key = mapKey(e);
    if (!key) return;
    // A key may have been pressed while the page had focus and released after
    // focus moved to a settings control. Always release our own held token;
    // otherwise the desktop can retain a stuck modifier/key.
    const interactive = core?.isInteractiveTarget(e.target);
    if (!interactive) e.preventDefault();
    releaseInput(keyHolds, key, `hardware:${e.code || e.key}`, 'key');
  });

  // Gamepad profiles, adaptive-centre sticks and layout editor.
  function safeStorageGet(key) {
    try { return localStorage.getItem(key); } catch { return null; }
  }
  function safeStorageSet(key, value) {
    try { localStorage.setItem(key, value); } catch { /* Private mode may disable storage. */ }
  }
  function saveLayoutFor(profileId, controls) {
    safeStorageSet(`touchbridge.layout.${profileId}`, JSON.stringify(controls));
  }
  function saveLayout() {
    saveLayoutFor(activeProfile, layout);
  }
  function removeLayoutFor(profileId) {
    try { localStorage.removeItem(`touchbridge.layout.${profileId}`); } catch { /* Private mode may disable storage. */ }
  }
  function getCustomProfiles() {
    let parsed = [];
    try { parsed = JSON.parse(safeStorageGet(PROFILES_KEY) || '[]'); } catch { return []; }
    if (!Array.isArray(parsed)) return [];
    const seen = new Set();
    return parsed.flatMap(item => {
      if (!item || typeof item.id !== 'string' || !/^custom-[a-z0-9-]{1,56}$/.test(item.id)) return [];
      if (seen.has(item.id) || PRESET_CARDS[item.id]) return [];
      const name = typeof item.name === 'string' ? item.name.trim().slice(0, 30) : '';
      if (!name) return [];
      seen.add(item.id);
      return [{ id: item.id, name }];
    });
  }
  function saveCustomProfiles(profiles) {
    safeStorageSet(PROFILES_KEY, JSON.stringify(profiles.map(({ id, name }) => ({ id, name }))));
  }
  function presetIds() {
    return Object.keys(core?.PROFILE_META || PRESET_CARDS);
  }
  function allProfiles() {
    const presets = presetIds().map(id => ({
      id,
      name: PRESET_CARDS[id]?.name || core?.PROFILE_META[id]?.name || id,
      subtitle: PRESET_CARDS[id]?.subtitle || 'Preset layout',
      preset: true
    }));
    const custom = getCustomProfiles().map(profile => ({ ...profile, subtitle: 'Custom layout', preset: false }));
    return [...presets, ...custom];
  }
  function profileExists(profileId) {
    return presetIds().includes(profileId) || getCustomProfiles().some(profile => profile.id === profileId);
  }
  function displayName(profileId) {
    return PRESET_CARDS[profileId]?.name
      || getCustomProfiles().find(profile => profile.id === profileId)?.name
      || 'Controller';
  }
  function setLayoutStatus(message) {
    if (layoutStatus) layoutStatus.textContent = message;
  }
  function layoutNameTaken(name, exceptId) {
    const needle = name.toLowerCase();
    return allProfiles().some(profile => profile.id !== exceptId && profile.name.toLowerCase() === needle);
  }
  function validateLayoutName(rawName, exceptId) {
    const name = String(rawName || '').trim().replace(/\s+/g, ' ').slice(0, 30);
    if (!name) return { error: 'Enter a layout name first.' };
    if (layoutNameTaken(name, exceptId)) return { error: 'A layout with that name already exists.' };
    return { name };
  }
  function makeCustomId() {
    return `custom-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 6)}`;
  }
  function renderHomeLayouts() {
    if (!profileGrid) return;
    const profiles = allProfiles();
    profileGrid.replaceChildren();
    for (const profile of profiles) {
      const card = document.createElement('article');
      card.className = 'profile-card';
      card.dataset.profileCard = profile.id;
      const art = document.createElement('div');
      art.className = `profile-art ${PRESET_CARDS[profile.id]?.art || 'generic'}-art`;
      art.setAttribute('aria-hidden', 'true');
      art.innerHTML = PROFILE_ART[PRESET_CARDS[profile.id]?.art] || PROFILE_ART.generic;
      const copy = document.createElement('div');
      copy.className = 'profile-card-copy';
      const text = document.createElement('div');
      const strong = document.createElement('strong');
      strong.textContent = profile.name;
      const small = document.createElement('small');
      small.textContent = profile.subtitle;
      text.append(strong, small);
      copy.append(text);
      const actions = document.createElement('div');
      actions.className = 'profile-card-actions';
      const play = document.createElement('button');
      play.type = 'button';
      play.dataset.playProfile = profile.id;
      play.textContent = 'Play';
      const edit = document.createElement('button');
      edit.type = 'button';
      edit.dataset.editProfile = profile.id;
      edit.textContent = 'Edit';
      actions.append(play, edit);
      card.append(art, copy, actions);
      profileGrid.append(card);
    }
    const customCount = profiles.filter(profile => !profile.preset).length;
    if (libraryCount) {
      libraryCount.textContent = customCount
        ? `${profiles.length - customCount} PRESETS · ${customCount} CUSTOM`
        : `${profiles.length} PRESETS`;
    }
  }
  function renderProfileOptions() {
    if (!profileSelect) return;
    profileSelect.replaceChildren();
    for (const profile of allProfiles()) {
      const option = document.createElement('option');
      option.value = profile.id;
      option.textContent = profile.name;
      profileSelect.append(option);
    }
    if (profileExists(activeProfile)) profileSelect.value = activeProfile;
  }
  function createLayout(rawName) {
    const result = validateLayoutName(rawName);
    if (result.error) { setLayoutStatus(result.error); newLayoutName?.focus(); return false; }
    const profiles = getCustomProfiles();
    const id = makeCustomId();
    profiles.push({ id, name: result.name });
    saveCustomProfiles(profiles);
    saveLayoutFor(id, core.createProfile('generic'));
    renderHomeLayouts();
    renderProfileOptions();
    renderLayoutManager();
    newLayoutForm.hidden = true;
    if (newLayoutButton) newLayoutButton.disabled = false;
    if (newLayoutName) newLayoutName.value = '';
    setLayoutStatus(`Created “${result.name}”. Open it from the library or edit it here.`);
    return true;
  }
  function duplicateLayout(id) {
    const source = allProfiles().find(profile => profile.id === id);
    if (!source || !core) return;
    const base = `${source.name} copy`.slice(0, 28);
    let name = base;
    let suffix = 2;
    while (layoutNameTaken(name)) { name = `${base} ${suffix++}`.slice(0, 30); }
    const profiles = getCustomProfiles();
    const newId = makeCustomId();
    profiles.push({ id: newId, name });
    saveCustomProfiles(profiles);
    saveLayoutFor(newId, loadLayout(id).map(control => ({ ...control })));
    renderHomeLayouts();
    renderProfileOptions();
    renderLayoutManager();
    setLayoutStatus(`Duplicated “${source.name}” as “${name}”.`);
  }
  function renameLayout(id, rawName) {
    if (presetIds().includes(id)) return;
    const result = validateLayoutName(rawName, id);
    if (result.error) { setLayoutStatus(result.error); return; }
    const profiles = getCustomProfiles().map(profile => profile.id === id ? { ...profile, name: result.name } : profile);
    saveCustomProfiles(profiles);
    renamingLayoutId = null;
    if (activeProfile === id) { updateProfileLabels(); renderProfileOptions(); }
    renderHomeLayouts();
    renderLayoutManager();
    setLayoutStatus(`Renamed to “${result.name}”.`);
  }
  function deleteLayout(id) {
    if (presetIds().includes(id)) { setLayoutStatus('Preset layouts cannot be deleted. Duplicate one to customize it.'); return; }
    const profile = getCustomProfiles().find(item => item.id === id);
    if (!profile) return;
    removeLayoutFor(id);
    saveCustomProfiles(getCustomProfiles().filter(item => item.id !== id));
    deletingLayoutId = null;
    if (renamingLayoutId === id) renamingLayoutId = null;
    if (activeProfile === id) {
      activeProfile = 'generic';
      safeStorageSet('touchbridge.lastProfile', activeProfile);
      layout = loadLayout(activeProfile);
      if (profileSelect) profileSelect.value = activeProfile;
      updateProfileLabels();
      renderGamepadLayout();
    }
    renderHomeLayouts();
    renderProfileOptions();
    renderLayoutManager();
    setLayoutStatus(`Deleted “${profile.name}”.`);
  }
  function layoutActionButton(label, action, danger = false) {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = `layout-action${danger ? ' danger-link' : ''}`;
    button.dataset.layoutAction = action;
    button.textContent = label;
    return button;
  }
  function renderLayoutManager() {
    if (!layoutList) return;
    const profiles = allProfiles();
    layoutList.replaceChildren();
    if (layoutManageCount) layoutManageCount.textContent = `${profiles.length} LAYOUTS`;
    for (const profile of profiles) {
      const row = document.createElement('div');
      row.className = 'layout-manage-row';
      row.dataset.layoutId = profile.id;
      const copy = document.createElement('div');
      copy.className = 'layout-manage-copy';
      const actions = document.createElement('div');
      actions.className = 'layout-manage-actions';
      if (deletingLayoutId === profile.id) {
        const strong = document.createElement('strong');
        strong.textContent = `Delete “${profile.name}”?`;
        const small = document.createElement('small');
        small.textContent = 'This removes its controls from this phone. This cannot be undone.';
        copy.append(strong, small);
        actions.append(layoutActionButton('Cancel', 'cancel-delete'), layoutActionButton('Delete', 'confirm-delete', true));
      } else if (renamingLayoutId === profile.id) {
        const input = document.createElement('input');
        input.className = 'layout-rename-input';
        input.type = 'text';
        input.maxLength = 30;
        input.value = profile.name;
        input.setAttribute('aria-label', 'Layout name');
        copy.append(input);
        actions.append(layoutActionButton('Save', 'save-rename'), layoutActionButton('Cancel', 'cancel-rename'));
      } else {
        const strong = document.createElement('strong');
        strong.textContent = profile.name;
        const small = document.createElement('small');
        const count = profileExists(profile.id) ? loadLayout(profile.id).length : 0;
        small.textContent = `${profile.preset ? 'Preset' : 'Custom'} · ${count} controls${profile.id === activeProfile ? ' · Active' : ''}`;
        copy.append(strong, small);
        actions.append(layoutActionButton('Edit', 'edit'));
        actions.append(layoutActionButton('Duplicate', 'duplicate'));
        if (!profile.preset) {
          actions.append(layoutActionButton('Rename', 'rename'));
          actions.append(layoutActionButton('Delete', 'delete', true));
        }
      }
      row.append(copy, actions);
      layoutList.append(row);
    }
    if (renamingLayoutId) {
      const input = layoutList.querySelector('.layout-rename-input');
      if (input) { input.focus(); input.select(); }
    }
  }
  function loadLayout(profileId) {
    const stored = safeStorageGet(`touchbridge.layout.${profileId}`);
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        if (core && core.validateLayout(parsed)) return parsed;
      } catch { /* Fall back to the built-in profile. */ }
    }
    return core ? core.createProfile(profileId) : [];
  }
  function selectedControl() {
    return layout.find(control => control.id === selectedControlId) || null;
  }
  function populateActionSelect(select) {
    select.replaceChildren();
    for (const action of core.BUTTONS) {
      const option = document.createElement('option');
      option.value = action;
      option.textContent = action;
      select.appendChild(option);
    }
  }
  function populateAddControlSelect() {
    if (!addControlAction || !core) return;
    addControlAction.replaceChildren();
    const singles = document.createElement('optgroup');
    singles.label = 'Single buttons';
    for (const action of core.BUTTONS) {
      const option = document.createElement('option');
      option.value = `button:${action}`;
      option.textContent = BUTTON_LABELS[action] || action;
      singles.appendChild(option);
    }
    const groups = document.createElement('optgroup');
    groups.label = 'Button sets';
    for (const [value, label] of [
      ['group:dpad', 'D-pad (4 buttons)'],
      ['group:face', 'Face buttons (4)'],
      ['group:shoulders', 'Bumpers (LB / RB)'],
      ['group:triggers', 'Triggers (LT / RT)']
    ]) {
      const option = document.createElement('option');
      option.value = value;
      option.textContent = label;
      groups.appendChild(option);
    }
    const racing = document.createElement('optgroup');
    racing.label = 'Racing controls';
    for (const [value, label] of [
      ['wheel', 'Steering wheel (left stick)'],
      ['pedal:RT', 'Accelerator pedal (analog RT)'],
      ['pedal:LT', 'Brake pedal (analog LT)']
    ]) {
      const option = document.createElement('option');
      option.value = value;
      option.textContent = label;
      racing.appendChild(option);
    }
    addControlAction.append(singles, groups, racing);
  }
  function updateEditorFields() {
    const control = selectedControl();
    selectedControlName.textContent = control ? (control.label || control.id).toUpperCase() : 'Select a control';
    if (layoutEditor) layoutEditor.hidden = !editMode || !control;
    controlLabel.disabled = !control;
    controlAction.disabled = !control || control.kind !== 'button';
    controlSize.disabled = !control;
    controlOpacity.disabled = !control;
    deleteControlButton.disabled = !control || control.kind === 'stick';
    if (controlList) {
      controlList.replaceChildren();
      for (const item of layout) {
        const button = document.createElement('button');
        button.type = 'button';
        button.className = `control-list-item${item.id === selectedControlId ? ' selected' : ''}`;
        button.dataset.controlId = item.id;
        const title = document.createElement('strong');
        title.textContent = item.label || item.id;
        const subtitle = document.createElement('small');
        subtitle.textContent = item.kind === 'stick'
          ? `${item.side.toUpperCase()} STICK`
          : item.kind === 'wheel' ? 'STEERING WHEEL' : (item.pedal ? `ANALOG ${item.button}` : item.button);
        button.append(title, subtitle);
        button.addEventListener('click', () => selectControl(item.id));
        controlList.append(button);
      }
    }
    if (!control) {
      controlLabel.value = '';
      controlSize.value = '46';
      controlSizeValue.value = '46 px';
      controlOpacity.value = '100';
      controlOpacityValue.value = '100%';
      return;
    }
    controlLabel.value = control.label;
    controlSize.value = String(control.size);
    controlSizeValue.value = `${control.size} px`;
    const opacity = Math.round((Number.isFinite(control.opacity) ? control.opacity : 1) * 100);
    controlOpacity.value = String(opacity);
    controlOpacityValue.value = `${opacity}%`;
    if (control.kind === 'button') controlAction.value = control.button;
  }
  function selectControl(id) {
    selectedControlId = id;
    gamepadControls.querySelectorAll('.layout-control').forEach(node => node.classList.toggle('selected', node.dataset.controlId === id));
    updateEditorFields();
  }
  function setControlPosition(control, node) {
    node.style.left = `${control.x}%`;
    node.style.top = `${control.y}%`;
    node.style.width = `${control.size}px`;
    node.style.height = `${control.size}px`;
    node.style.opacity = String(Number.isFinite(control.opacity) ? control.opacity : 1);
  }
  // Tapping empty canvas space clears the selection so the compact property
  // controls disappear and the editor exposes the same usable area as play.
  gamepadStage.addEventListener('pointerdown', event => {
    if (!editMode || event.target.closest('.layout-control')) return;
    selectedControlId = null;
    updateEditorFields();
  });
  // A long press on a control should not open the WebView context menu or add
  // its delayed native haptic on top of the immediate press feedback.
  gamepadStage.addEventListener('contextmenu', event => event.preventDefault());

  function renderGamepadLayout() {
    if (!core || !gamepadControls) return;
    gamepadControls.replaceChildren();
    gamepadStage.classList.toggle('editing', editMode);
    for (const control of layout) {
      const wrapper = document.createElement('div');
      wrapper.className = `layout-control ${control.kind === 'stick' ? 'stick-control' : control.kind === 'wheel' ? 'wheel-control' : 'button-control'}`;
      wrapper.dataset.controlId = control.id;
      setControlPosition(control, wrapper);
      if (control.kind === 'wheel') {
        const face = document.createElement('div');
        face.className = 'wheel-face';
        const rim = document.createElement('div');
        rim.className = 'wheel-rim';
        const hub = document.createElement('div');
        hub.className = 'wheel-hub';
        const mark = document.createElement('div');
        mark.className = 'wheel-mark';
        const tag = document.createElement('div');
        tag.className = 'stick-tag wheel-tag';
        tag.textContent = control.label || 'WHEEL';
        rim.append(hub, mark);
        face.appendChild(rim);
        face.appendChild(tag);
        wrapper.appendChild(face);
        wrapper._wheelParts = { face, rim, tag };
      } else if (control.kind === 'stick') {
        const face = document.createElement('div');
        face.className = 'stick-face';
        const center = document.createElement('div');
        center.className = 'stick-center';
        const ring = document.createElement('div');
        ring.className = 'stick-origin-ring';
        const knob = document.createElement('div');
        knob.className = 'stick-knob';
        const tag = document.createElement('div');
        tag.className = 'stick-tag';
        tag.textContent = control.label || (control.side === 'left' ? 'LEFT STICK' : 'RIGHT STICK');
        // Keep the origin marker above the knob: otherwise the knob hides the
        // adaptive centre exactly when the finger first touches the stick.
        face.append(knob, center, ring, tag);
        wrapper.appendChild(face);
        wrapper._stickParts = { center, ring, knob, face };
      } else {
        const button = document.createElement('button');
        button.type = 'button';
        button.className = 'gp-button';
        button.textContent = control.label;
        button.dataset.action = control.button;
        button.title = control.hint || `${control.label} · ${control.button}`;
        if (control.button.startsWith('DPAD_')) button.classList.add('dpad-button');
        if (control.button === 'LT' || control.button === 'RT' || control.button === 'LB' || control.button === 'RB') button.classList.add('trigger-button');
        if (control.pedal) {
          wrapper.classList.add('pedal-control');
          button.classList.add('pedal-button');
        }
        if (control.button === 'START' || control.button === 'BACK' || control.button === 'GUIDE') button.classList.add('small-action');
        wrapper.appendChild(button);
      }
      wrapper.addEventListener('pointerdown', event => onGamepadPointerDown(event, control, wrapper));
      wrapper.addEventListener('pointermove', event => onGamepadPointerMove(event, control, wrapper));
      wrapper.addEventListener('pointerup', event => onGamepadPointerEnd(event, control, wrapper, true));
      wrapper.addEventListener('pointercancel', event => onGamepadPointerEnd(event, control, wrapper, false));
      wrapper.addEventListener('lostpointercapture', event => onGamepadPointerEnd(event, control, wrapper, false));
      wrapper.addEventListener('click', () => { if (editMode) selectControl(control.id); });
      gamepadControls.appendChild(wrapper);
      if (control.id === selectedControlId) wrapper.classList.add('selected');
    }
    updateEditorFields();
    refreshGyroStatus();
  }
  function setGamepadAvailability(available, message) {
    gamepadAvailable = Boolean(available);
    updateAboutDiagnostics();
    gamepadSupport.classList.toggle('available', gamepadAvailable);
    gamepadSupport.classList.toggle('unavailable', !gamepadAvailable && authenticated);
    if (gamepadAvailable) {
      gamepadSupport.textContent = message || 'Virtual gamepad is available on this PC. Select a profile and start playing.';
    } else {
      gamepadSupport.textContent = message || (authenticated ? 'This PC does not currently expose a virtual gamepad device. Mouse and keyboard still work.' : 'Connect to a PC to check virtual gamepad support. Mouse and keyboard remain available.');
      clearGamepadState(false);
    }
  }
  function scheduleGamepadFlush(delay = 0) {
    if (!gamepadDirty || !authenticated || !gamepadAvailable || gamepadFrame || gamepadRetryTimer) return;
    if (delay > 0) {
      gamepadRetryTimer = window.setTimeout(() => {
        gamepadRetryTimer = 0;
        scheduleGamepadFlush();
      }, delay);
    } else {
      gamepadFrame = requestAnimationFrame(flushGamepadState);
    }
  }
  function flushGamepadState() {
    gamepadFrame = 0;
    if (!gamepadDirty || !authenticated || !gamepadAvailable) {
      gamepadDirty = false;
      return;
    }
    if (socket && socket.bufferedAmount > 32768) {
      scheduleGamepadFlush(20);
      return;
    }
    const sent = send({
      type: 'gamepad_state',
      left_x: gamepadState.left_x, left_y: gamepadState.left_y,
      right_x: gamepadState.right_x, right_y: gamepadState.right_y,
      left_trigger: gamepadState.left_trigger, right_trigger: gamepadState.right_trigger,
      buttons: Array.from(gamepadButtons)
    });
    if (sent) gamepadDirty = false;
    else scheduleGamepadFlush(20);
  }
  function queueGamepadState() {
    if (!authenticated || !gamepadAvailable) return;
    gamepadDirty = true;
    scheduleGamepadFlush();
  }
  function isTriggerAction(action) {
    return action === 'LT' || action === 'RT';
  }
  function triggerValue(control, clientY, rect) {
    return control.pedal
      ? core.pedalPressure(clientY, rect.top, rect.height)
      : core.triggerPressure(clientY, rect.top, rect.height);
  }
  function rebuildGamepadButtons() {
    gamepadButtons.clear();
    gamepadState.left_trigger = 0;
    gamepadState.right_trigger = 0;
    for (const held of activeGamepadControls.values()) {
      const action = held.action;
      if (isTriggerAction(action)) {
        if (held.analog) {
          if (action === 'LT') gamepadState.left_trigger = Math.max(gamepadState.left_trigger, held.value);
          else gamepadState.right_trigger = Math.max(gamepadState.right_trigger, held.value);
        } else {
          // Digital mode reports the button as well as full pressure. Some
          // platforms (notably Linux evdev) expose LT/RT as both a button and
          // an axis; sending only the button left the axis at zero in games.
          gamepadButtons.add(action);
          if (action === 'LT') gamepadState.left_trigger = 1;
          else gamepadState.right_trigger = 1;
        }
      } else gamepadButtons.add(action);
    }
    queueGamepadState();
  }

  function updateGamepadPressedClasses() {
    if (!gamepadControls) return;
    gamepadControls.querySelectorAll('.layout-control').forEach(node => {
      const active = [...activeGamepadControls.values()].filter(item => item.controlId === node.dataset.controlId);
      const button = node.querySelector('.gp-button');
      if (button) {
        button.classList.toggle('pressed', active.length > 0);
        if (button.dataset.action === 'LT' || button.dataset.action === 'RT') {
          const pressure = active.reduce((max, item) => Math.max(max, item.value), 0);
          button.style.setProperty('--trigger-pressure', `${pressure * 100}%`);
        }
      }
    });
  }
  function clearGamepadState(sendReset, sendUpdate = true) {
    stickGestures.clear();
    activeGamepadControls.clear();
    gamepadButtons.clear();
    wheelGesture = null;
    updateWheelVisual(0);
    if (gamepadControls) gamepadControls.querySelectorAll('.active-wheel').forEach(node => node.classList.remove('active-wheel'));
    Object.assign(gamepadState, { left_x: 0, left_y: 0, right_x: 0, right_y: 0, left_trigger: 0, right_trigger: 0 });
    if (gamepadFrame) cancelAnimationFrame(gamepadFrame);
    if (gamepadRetryTimer) clearTimeout(gamepadRetryTimer);
    gamepadFrame = gamepadRetryTimer = 0;
    gamepadDirty = false;
    if (sendReset && authenticated && gamepadAvailable) send({ type: 'gamepad_reset' });
    else if (sendUpdate && authenticated && gamepadAvailable) queueGamepadState();
    if (gamepadControls) {
      gamepadControls.querySelectorAll('.stick-center,.stick-origin-ring,.stick-knob').forEach(node => {
        node.style.left = '50%';
        node.style.top = '50%';
      });
      gamepadControls.querySelectorAll('.active-stick').forEach(node => node.classList.remove('active-stick'));
      gamepadControls.querySelectorAll('.pressed').forEach(node => node.classList.remove('pressed'));
    }
  }
  function triggerHaptic() {
    if (!hapticsEnabled?.checked) return;
    try {
      if (window.TouchBridgeAndroid?.performHapticFeedback) window.TouchBridgeAndroid.performHapticFeedback();
      else if (typeof navigator.vibrate === 'function') navigator.vibrate(16);
    } catch { /* Haptics must never block controller input. */ }
  }
  function onGamepadPointerDown(event, control, wrapper) {
    if (event.button !== 0 && event.pointerType === 'mouse') return;
    event.preventDefault();
    if (editMode) {
      selectControl(control.id);
      const rect = gamepadStage.getBoundingClientRect();
      if (!(rect.width > 0 && rect.height > 0)) return;
      layoutDrag = { pointerId: event.pointerId, controlId: control.id, startX: event.clientX, startY: event.clientY, originX: control.x, originY: control.y, stageWidth: rect.width, stageHeight: rect.height };
      wrapper.setPointerCapture(event.pointerId);
      return;
    }
    if (control.kind === 'button') triggerHaptic();
    // Keep the controller tactile in offline preview mode. Input snapshots are
    // sent only when an authenticated server exposes a gamepad backend.
    wrapper.setPointerCapture(event.pointerId);
    if (control.kind === 'wheel') {
      const rect = wrapper.getBoundingClientRect();
      wheelGesture = {
        pointerId: event.pointerId,
        controlId: control.id,
        wrapper,
        angle: Math.atan2(event.clientY - (rect.top + rect.height / 2), event.clientX - (rect.left + rect.width / 2)),
        rotation: 0,
        maxRotation: WHEEL_MAX_ROTATION
      };
      wrapper.classList.add('active-wheel');
      return;
    }
    if (control.kind === 'stick') {
      const rect = wrapper.getBoundingClientRect();
      const width = rect.width;
      const height = rect.height;
      const localX = event.clientX - rect.left;
      const localY = event.clientY - rect.top;
      const origin = centerModeSelect.value === 'adaptive'
        ? core.adaptiveOrigin(localX, localY, width, height, Number(centerRange.value) / 100)
        : { x: width / 2, y: height / 2 };
      stickGestures.set(control.side, {
        pointerId: event.pointerId,
        wrapper,
        side: control.side,
        originX: origin.x,
        originY: origin.y,
        width,
        height,
        maxThrow: Math.min(width, height) * 0.39,
        center: wrapper._stickParts.center,
        knob: wrapper._stickParts.knob
      });
      wrapper.classList.add('active-stick');
      for (const part of [wrapper._stickParts.center, wrapper._stickParts.ring]) {
        part.style.left = `${origin.x}px`;
        part.style.top = `${origin.y}px`;
      }
      updateStickGesture(control.side, event.clientX, event.clientY);
    } else {
      const rect = wrapper.getBoundingClientRect();
      const isTrigger = isTriggerAction(control.button);
      const analog = isTrigger && (control.pedal || triggerModeSelect.value === 'analog');
      const value = analog ? triggerValue(control, event.clientY, rect) : 1;
      activeGamepadControls.set(event.pointerId, { controlId: control.id, action: control.button, value, analog });
      updateGamepadPressedClasses();
      rebuildGamepadButtons();
    }
  }

  function updateWheelVisual(rotation) {
    const wheel = gamepadControls.querySelector('.wheel-control');
    if (!wheel) return;
    const rim = wheel.querySelector('.wheel-rim');
    if (rim) rim.style.transform = `translate(-50%,-50%) rotate(${rotation * 180 / Math.PI}deg)`;
  }

  function setWheelAxis(axis) {
    gamepadState.left_x = axis;
    updateWheelVisual(axis * WHEEL_MAX_ROTATION);
    queueGamepadState();
  }
  function updateStickGesture(side, clientX, clientY) {
    const gesture = stickGestures.get(side);
    if (!gesture) return;
    const rect = gesture.wrapper.getBoundingClientRect();
    const localX = clientX - rect.left;
    const localY = clientY - rect.top;
    let dx = localX - gesture.originX;
    let dy = localY - gesture.originY;
    const distance = Math.hypot(dx, dy);
    if (distance > gesture.maxThrow) {
      dx = dx / distance * gesture.maxThrow;
      dy = dy / distance * gesture.maxThrow;
    }
    const deadzone = Number(stickDeadzone.value) / 100;
    const vector = core.normalizeStick(dx / gesture.maxThrow, -dy / gesture.maxThrow, deadzone, stickCurve.value);
    gamepadState[`${side}_x`] = vector.x;
    gamepadState[`${side}_y`] = vector.y;
    gesture.knob.style.left = `${gesture.originX + dx}px`;
    gesture.knob.style.top = `${gesture.originY + dy}px`;
    queueGamepadState();
  }
  function onGamepadPointerMove(event, control, wrapper) {
    if (editMode && layoutDrag && layoutDrag.pointerId === event.pointerId) {
      event.preventDefault();
      const drag = layoutDrag;
      const item = layout.find(candidate => candidate.id === drag.controlId);
      if (!item) return;
      const rawX = drag.originX + (event.clientX - drag.startX) / drag.stageWidth * 100;
      const rawY = drag.originY + (event.clientY - drag.startY) / drag.stageHeight * 100;
      item.x = core.snapPercent(rawX, editorSnapEnabled, 2);
      item.y = core.snapPercent(rawY, editorSnapEnabled, 2);
      setControlPosition(item, wrapper);
      return;
    }
    if (control.kind === 'wheel') {
      if (!wheelGesture || wheelGesture.pointerId !== event.pointerId) return;
      event.preventDefault();
      const rect = wrapper.getBoundingClientRect();
      const angle = Math.atan2(event.clientY - (rect.top + rect.height / 2), event.clientX - (rect.left + rect.width / 2));
      let delta = angle - wheelGesture.angle;
      if (delta > Math.PI) delta -= Math.PI * 2;
      if (delta < -Math.PI) delta += Math.PI * 2;
      wheelGesture.angle = angle;
      wheelGesture.rotation = core.clamp(wheelGesture.rotation + delta, -wheelGesture.maxRotation, wheelGesture.maxRotation);
      setWheelAxis(core.wheelAxis(wheelGesture.rotation, wheelGesture.maxRotation));
      return;
    }
    if (control.kind === 'button') {
      const held = activeGamepadControls.get(event.pointerId);
      if (!held || !isTriggerAction(held.action) || !held.analog) return;
      event.preventDefault();
      const rect = wrapper.getBoundingClientRect();
      held.value = triggerValue(control, event.clientY, rect);
      updateGamepadPressedClasses();
      rebuildGamepadButtons();
      return;
    }
    const gesture = stickGestures.get(control.side);
    if (!gesture || gesture.pointerId !== event.pointerId) return;
    event.preventDefault();
    const samples = typeof event.getCoalescedEvents === 'function' ? event.getCoalescedEvents() : [event];
    const current = samples.length ? samples[samples.length - 1] : event;
    updateStickGesture(control.side, current.clientX, current.clientY);
  }
  function onGamepadPointerEnd(event, control, wrapper, wasUp) {
    if (layoutDrag && layoutDrag.pointerId === event.pointerId) {
      layoutDrag = null;
      saveLayout();
      return;
    }
    if (control.kind === 'wheel') {
      if (!wheelGesture || wheelGesture.pointerId !== event.pointerId) return;
      wheelGesture.wrapper.classList.remove('active-wheel');
      wheelGesture = null;
      // A wheel self-centres when released, like a real steering wheel.
      if (!stickGestures.has('left')) setWheelAxis(0);
      return;
    }
    if (control.kind === 'stick') {
      const gesture = stickGestures.get(control.side);
      if (!gesture || gesture.pointerId !== event.pointerId) return;
      stickGestures.delete(control.side);
      gamepadState[`${control.side}_x`] = 0;
      gamepadState[`${control.side}_y`] = 0;
      gesture.center.style.left = '50%';
      gesture.center.style.top = '50%';
      gesture.knob.style.left = '50%';
      gesture.knob.style.top = '50%';
      const ring = gesture.wrapper._stickParts?.ring;
      if (ring) { ring.style.left = '50%'; ring.style.top = '50%'; }
      gesture.wrapper.classList.remove('active-stick');
      queueGamepadState();
    } else if (activeGamepadControls.has(event.pointerId)) {
      activeGamepadControls.delete(event.pointerId);
      updateGamepadPressedClasses();
      rebuildGamepadButtons();
    }
  }
  function updateEditorToggles() {
    gamepadStage.classList.toggle('show-grid', editorGridEnabled && editMode);
    if (editorGridButton) {
      editorGridButton.setAttribute('aria-pressed', String(editorGridEnabled));
      editorGridButton.title = editorGridEnabled ? 'Hide grid' : 'Show grid';
    }
    if (editorSnapButton) {
      editorSnapButton.setAttribute('aria-pressed', String(editorSnapEnabled));
      editorSnapButton.title = editorSnapEnabled ? 'Disable snap' : 'Enable snap';
    }
  }
  function findOpenControlPosition(size) {
    const rect = gamepadStage.getBoundingClientRect();
    return core.findOpenPosition(layout, size, rect.width, rect.height);
  }
  function setEditorMode(enabled) {
    enabled = Boolean(enabled);
    // Never leave a virtual stick/button held while the layout DOM is rebuilt.
    if (enabled && (stickGestures.size || activeGamepadControls.size)) {
      clearGamepadState(true, false);
    }
    editMode = enabled;
    layoutEditor.hidden = !editMode || !selectedControl();
    if (editorAdvancedPanel) editorAdvancedPanel.hidden = true;
    if ($('editorMoreButton')) $('editorMoreButton').setAttribute('aria-expanded', 'false');
    updateEditorToggles();
    if (editLayoutButton) editLayoutButton.textContent = 'Edit';
    selectedControlId = null;
    renderGamepadLayout();
  }

  populateActionSelect(controlAction);
  populateAddControlSelect();
  if (core) {
    const savedProfile = safeStorageGet('touchbridge.lastProfile');
    activeProfile = savedProfile && profileExists(savedProfile) ? savedProfile : 'generic';
    renderProfileOptions();
    profileSelect.value = activeProfile;
    layout = loadLayout(activeProfile);
    updateProfileLabels();
    renderGamepadLayout();
    renderHomeLayouts();
    renderLayoutManager();
  }
  profileSelect.addEventListener('change', () => setProfile(profileSelect.value));
  editLayoutButton?.addEventListener('click', openEditor);
  editorGridButton.addEventListener('click', () => {
    editorGridEnabled = !editorGridEnabled;
    safeStorageSet('touchbridge.editor.grid', String(editorGridEnabled));
    updateEditorToggles();
  });
  editorSnapButton.addEventListener('click', () => {
    editorSnapEnabled = !editorSnapEnabled;
    safeStorageSet('touchbridge.editor.snap', String(editorSnapEnabled));
    updateEditorToggles();
  });
  controlLabel.addEventListener('input', () => {
    const control = selectedControl();
    if (!control) return;
    control.label = controlLabel.value.slice(0, 12);
    const wrapper = gamepadControls.querySelector(`[data-control-id="${CSS.escape(control.id)}"]`);
    if (wrapper) {
      const button = wrapper.querySelector('.gp-button');
      const tag = wrapper.querySelector('.stick-tag');
      if (button) button.textContent = control.label;
      if (tag) tag.textContent = control.label;
    }
    saveLayout();
  });
  controlAction.addEventListener('change', () => {
    const control = selectedControl();
    if (!control || control.kind !== 'button') return;
    control.button = controlAction.value;
    if (control.pedal && !isTriggerAction(control.button)) delete control.pedal;
    saveLayout();
    renderGamepadLayout();
    selectControl(control.id);
  });
  controlSize.addEventListener('input', () => {
    const control = selectedControl();
    if (!control) return;
    control.size = Number(controlSize.value);
    controlSizeValue.value = `${control.size} px`;
    const opacity = Math.round((Number.isFinite(control.opacity) ? control.opacity : 1) * 100);
    controlOpacity.value = String(opacity);
    controlOpacityValue.value = `${opacity}%`;
    const wrapper = gamepadControls.querySelector(`[data-control-id="${CSS.escape(control.id)}"]`);
    if (wrapper) setControlPosition(control, wrapper);
    saveLayout();
  });
  controlOpacity.addEventListener('input', () => {
    const control = selectedControl();
    if (!control) return;
    control.opacity = Number(controlOpacity.value) / 100;
    controlOpacityValue.value = `${controlOpacity.value}%`;
    const wrapper = gamepadControls.querySelector(`[data-control-id="${CSS.escape(control.id)}"]`);
    if (wrapper) wrapper.style.opacity = String(control.opacity);
    saveLayout();
  });
  let customControlSequence = 0;
  addControlButton.addEventListener('click', () => {
    const value = addControlAction.value || 'button:A';
    const isSet = value.startsWith('group:');
    const isPedal = value.startsWith('pedal:');
    const isWheel = value === 'wheel';
    const specs = isSet
      ? ADD_CLUSTERS[value.slice(6)]
      : isWheel
        ? [{ wheel: true, label: 'WHEEL', dx: 0, dy: 0, size: 150 }]
        : isPedal
          ? [{ button: value.slice(6), label: value === 'pedal:RT' ? 'ACCEL' : 'BRAKE', pedal: true, dx: 0, dy: 0, size: 88 }]
          : (() => {
            const action = value.startsWith('button:') ? value.slice(7) : value;
            return [{ button: action, label: BUTTON_LABELS[action] || action, dx: 0, dy: 0, size: 46 }];
          })();
    if (!Array.isArray(specs) || !specs.length) return;
    const remaining = 40 - layout.length;
    if (specs.length > remaining) {
      editorStatus.textContent = `This layout has room for ${Math.max(0, remaining)} more control${remaining === 1 ? '' : 's'}. Remove some before adding a set.`;
      return;
    }
    const position = findOpenControlPosition(Math.max(...specs.map(spec => spec.size || 46)));
    const added = [];
    for (const spec of specs) {
      const control = {
        id: `custom-${Date.now()}-${++customControlSequence}`,
        kind: spec.wheel ? 'wheel' : 'button',
        ...(spec.wheel ? {} : { button: spec.button }),
        ...(spec.pedal ? { pedal: true } : {}),
        label: spec.label,
        x: core.clamp(position.x + spec.dx, 4, 96),
        y: core.clamp(position.y + spec.dy, 4, 96),
        size: spec.size,
        opacity: 0.92
      };
      layout.push(control);
      added.push(control.id);
    }
    selectedControlId = added[added.length - 1];
    saveLayout();
    renderGamepadLayout();
    selectControl(selectedControlId);
    editorStatus.textContent = added.length > 1
      ? `Added ${added.length} controls. Drag them into place.`
      : 'Added one control. Drag it into place.';
  });
  deleteControlButton.addEventListener('click', () => {
    const control = selectedControl();
    if (!control || control.kind === 'stick') return;
    layout = layout.filter(item => item.id !== control.id);
    selectedControlId = null;
    saveLayout();
    renderGamepadLayout();
  });
  resetLayoutButton.addEventListener('click', () => {
    if (stickGestures.size || activeGamepadControls.size) clearGamepadState(true, false);
    layout = core.createProfile(activeProfile);
    selectedControlId = null;
    saveLayout();
    renderGamepadLayout();
    editorStatus.textContent = presetIds().includes(activeProfile)
      ? 'Layout restored to its starter profile.'
      : 'Layout reset to the Xbox-style starter controls.';
  });
  exportLayoutButton.addEventListener('click', () => {
    if (!core.validateLayout(layout)) {
      if (editorStatus) editorStatus.textContent = 'This layout is invalid and cannot be exported.';
      return;
    }
    const payload = {
      format: 'touchbridge-layout',
      version: 1,
      profile: activeProfile,
      exportedAt: new Date().toISOString(),
      controls: layout
    };
    const contents = JSON.stringify(payload, null, 2);
    const slug = displayName(activeProfile).toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'layout';
    const filename = `touchbridge-${slug}-layout.json`;
    if (window.TouchBridgeAndroid?.saveTextFile) {
      const queued = window.TouchBridgeAndroid.saveTextFile(filename, contents);
      if (editorStatus) editorStatus.textContent = queued ? 'Choose where to save this layout…' : 'The layout could not be exported.';
      return;
    }
    const blob = new Blob([contents], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    if (editorStatus) editorStatus.textContent = 'Layout exported as a JSON file.';
  });
  importLayoutButton.addEventListener('click', () => importLayoutFile.click());
  importLayoutFile.addEventListener('change', async () => {
    const file = importLayoutFile.files && importLayoutFile.files[0];
    importLayoutFile.value = '';
    if (!file) return;
    if (file.size > 64 * 1024) {
      if (editorStatus) editorStatus.textContent = 'The layout file is too large. Maximum size is 64 KB.';
      return;
    }
    try {
      const parsed = JSON.parse(await file.text());
      const imported = Array.isArray(parsed) ? parsed : parsed && parsed.controls;
      if (!core.validateLayout(imported)) throw new Error('Layout structure is invalid.');
      if (stickGestures.size || activeGamepadControls.size) clearGamepadState(true, false);
      layout = JSON.parse(JSON.stringify(imported));
      selectedControlId = null;
      saveLayout();
      renderGamepadLayout();
      if (editorStatus) editorStatus.textContent = `Imported ${layout.length} controls into the ${activeProfile} profile.`;
    } catch (error) {
      if (editorStatus) editorStatus.textContent = `Could not import layout: ${error.message || 'invalid JSON'}`;
    }
  });
  controlSize.addEventListener('change', updateEditorFields);
  centerRange.addEventListener('input', () => {
    centerRangeValue.value = `${centerRange.value}%`;
    safeStorageSet('touchbridge.stick.centerRange', centerRange.value);
  });
  stickDeadzone.addEventListener('input', () => {
    stickDeadzoneValue.value = `${stickDeadzone.value}%`;
    safeStorageSet('touchbridge.stick.deadzone', stickDeadzone.value);
  });
  centerModeSelect.addEventListener('change', () => safeStorageSet('touchbridge.stick.centerMode', centerModeSelect.value));
  stickCurve.addEventListener('change', () => safeStorageSet('touchbridge.stick.curve', stickCurve.value));
  const savedGrid = safeStorageGet('touchbridge.editor.grid');
  if (savedGrid === 'true' || savedGrid === 'false') editorGridEnabled = savedGrid === 'true';
  const savedSnap = safeStorageGet('touchbridge.editor.snap');
  if (savedSnap === 'true' || savedSnap === 'false') editorSnapEnabled = savedSnap === 'true';
  updateEditorToggles();
  const savedCenterMode = safeStorageGet('touchbridge.stick.centerMode');
  if (savedCenterMode === 'adaptive' || savedCenterMode === 'fixed') centerModeSelect.value = savedCenterMode;
  const savedRange = Number(safeStorageGet('touchbridge.stick.centerRange'));
  if (Number.isFinite(savedRange) && savedRange >= 20 && savedRange <= 65) centerRange.value = String(savedRange);
  centerRangeValue.value = `${centerRange.value}%`;
  const savedDeadzone = Number(safeStorageGet('touchbridge.stick.deadzone'));
  if (Number.isFinite(savedDeadzone) && savedDeadzone >= 0 && savedDeadzone <= 25) stickDeadzone.value = String(savedDeadzone);
  stickDeadzoneValue.value = `${stickDeadzone.value}%`;
  const savedCurve = safeStorageGet('touchbridge.stick.curve');
  if (['linear', 'precision', 'responsive'].includes(savedCurve)) stickCurve.value = savedCurve;
  hapticsEnabled.checked = safeStorageGet('touchbridge.haptics') !== 'false';
  const savedTheme = safeStorageGet('touchbridge.theme');
  if (themeSelect) {
    themeSelect.value = ['light', 'dark', 'amoled'].includes(savedTheme) ? savedTheme : 'system';
    applyThemePreference(themeSelect.value);
    themeSelect.addEventListener('change', () => applyThemePreference(themeSelect.value));
  } else {
    applyThemePreference(savedTheme);
  }
  const appVersion = window.TouchBridgeAndroid?.getAppVersion?.()
    || document.querySelector('meta[name="touchbridge-version"]')?.content
    || '0.6.0';
  if (appVersionValue) appVersionValue.textContent = `Version ${appVersion}`;
  $('openRepositoryButton')?.addEventListener('click', () => {
    if (window.TouchBridgeAndroid?.openRepository) window.TouchBridgeAndroid.openRepository();
    else window.open('https://github.com/Spont-427/touchbridge-controller', '_blank', 'noopener');
  });
  const themeQuery = window.matchMedia ? window.matchMedia('(prefers-color-scheme: light)') : null;
  if (themeQuery) {
    const onSchemeChange = () => {
      if (themeSelect.value === 'system') applyThemePreference('system');
    };
    if (themeQuery.addEventListener) themeQuery.addEventListener('change', onSchemeChange);
    else if (themeQuery.addListener) themeQuery.addListener(onSchemeChange);
  }
  const savedTriggerMode = safeStorageGet('touchbridge.triggerMode');
  triggerModeSelect.value = savedTriggerMode === 'digital' ? 'digital' : 'analog';
  const savedGyro = safeStorageGet('touchbridge.gyro.enabled');
  const savedGyroInvert = safeStorageGet('touchbridge.gyro.invert');
  if (gyroInvert) gyroInvert.checked = savedGyroInvert === 'true';
  refreshGyroStatus();
  if (savedGyro === 'true') setGyroEnabled(true);
  hapticsEnabled.addEventListener('change', () => safeStorageSet('touchbridge.haptics', String(hapticsEnabled.checked)));
  triggerModeSelect.addEventListener('change', () => {
    safeStorageSet('touchbridge.triggerMode', triggerModeSelect.value);
    rebuildGamepadButtons();
    updateGamepadPressedClasses();
    queueGamepadState();
  });

  function gyroSupported() {
    return typeof window.DeviceOrientationEvent !== 'undefined';
  }

  function currentScreenAngle() {
    const angle = window.screen?.orientation?.angle;
    if (Number.isFinite(angle)) return Number(angle);
    const legacy = Number(window.orientation);
    return Number.isFinite(legacy) ? legacy : 0;
  }

  function tiltDegrees(event) {
    const beta = Number(event.beta);
    const gamma = Number(event.gamma);
    if (!Number.isFinite(beta) && !Number.isFinite(gamma)) return null;
    const angle = ((currentScreenAngle() % 360) + 360) % 360;
    if (angle === 90) return -beta;
    if (angle === 270) return beta;
    if (angle === 180) return -gamma;
    return gamma;
  }

  function setGyroStatus(message) {
    if (gyroStatus) gyroStatus.textContent = message;
  }

  function refreshGyroStatus() {
    if (!gyroSteeringEnabled) return;
    if (!gyroSupported()) {
      gyroSteeringEnabled.checked = false;
      gyroSteeringEnabled.disabled = true;
      setGyroStatus('This device does not expose motion sensors to the controller.');
      return;
    }
    if (!gyroSteeringEnabled.checked) {
      setGyroStatus('Gyro steering is off. Add a steering wheel control to use it.');
      return;
    }
    if (!layout.some(control => control.kind === 'wheel')) {
      setGyroStatus('Enabled. Add a steering wheel to this layout to steer by tilting.');
      return;
    }
    if (!gyroListening) {
      setGyroStatus('Preparing the motion sensor…');
      return;
    }
    setGyroStatus(gyroNeutral === null
      ? 'Hold the phone in its level driving position to calibrate.'
      : 'Gyro steering is active. Calibrate again if the neutral point drifts.');
  }

  function onDeviceOrientation(event) {
    if (!gyroSteeringEnabled?.checked || currentScreen !== 'controller') return;
    if (!layout.some(control => control.kind === 'wheel')) return;
    if (wheelGesture || stickGestures.has('left')) return;
    const tilt = tiltDegrees(event);
    if (tilt === null) return;
    if (gyroNeutral === null) {
      gyroNeutral = tilt;
      gyroFiltered = 0;
      refreshGyroStatus();
      return;
    }
    let delta = tilt - gyroNeutral;
    if (gyroInvert?.checked) delta = -delta;
    // A light low-pass filter avoids visible wheel jitter without adding lag.
    gyroFiltered = gyroFiltered * 0.72 + delta * 0.28;
    const axis = core ? core.steeringFromTilt(gyroFiltered, 0, GYRO_MAX_TILT, 1.5) : 0;
    gamepadState.left_x = axis;
    updateWheelVisual(axis * WHEEL_MAX_ROTATION);
    queueGamepadState();
  }

  function attachGyro() {
    if (gyroListening || !gyroSupported()) return;
    window.addEventListener('deviceorientation', onDeviceOrientation);
    gyroListening = true;
    refreshGyroStatus();
  }

  function stopGyro() {
    if (!gyroListening) return;
    window.removeEventListener('deviceorientation', onDeviceOrientation);
    gyroListening = false;
    gyroNeutral = null;
    gyroFiltered = 0;
    if (gamepadState.left_x !== 0 && !stickGestures.has('left')) {
      gamepadState.left_x = 0;
      updateWheelVisual(0);
      queueGamepadState();
    }
  }

  function setGyroEnabled(enabled) {
    if (gyroSteeringEnabled) gyroSteeringEnabled.checked = Boolean(enabled);
    safeStorageSet('touchbridge.gyro.enabled', String(Boolean(enabled)));
    if (!enabled) { stopGyro(); refreshGyroStatus(); return; }
    if (!gyroSupported()) { safeStorageSet('touchbridge.gyro.enabled', 'false'); refreshGyroStatus(); return; }
    const requestPermission = window.DeviceOrientationEvent?.requestPermission;
    if (typeof requestPermission === 'function') {
      Promise.resolve(requestPermission.call(window.DeviceOrientationEvent)).then(result => {
        if (result === 'granted') attachGyro();
        else { gyroSteeringEnabled.checked = false; safeStorageSet('touchbridge.gyro.enabled', 'false'); setGyroStatus('Motion access was denied. Allow motion access to use gyro steering.'); }
      }).catch(() => {
        gyroSteeringEnabled.checked = false;
        safeStorageSet('touchbridge.gyro.enabled', 'false');
        setGyroStatus('Motion access is unavailable on this device.');
      });
      return;
    }
    attachGyro();
  }

  gyroSteeringEnabled?.addEventListener('change', () => setGyroEnabled(gyroSteeringEnabled.checked));
  gyroInvert?.addEventListener('change', () => safeStorageSet('touchbridge.gyro.invert', String(gyroInvert.checked)));
  gyroCalibrateButton?.addEventListener('click', () => {
    gyroNeutral = null;
    gyroFiltered = 0;
    setGyroStatus('Hold the phone in its level driving position to calibrate.');
    if (gyroSteeringEnabled?.checked && !gyroListening) attachGyro();
  });

  // The keyboard remains usable even when this PC has no virtual gamepad
  // provider. Linux uses uinput when permissions permit; Windows uses ViGEmBus.
  setGamepadAvailability(false, 'Connect to a PC to check virtual gamepad support. Mouse and keyboard remain available.');

  window.addEventListener('blur', () => {
    send({ type: 'release_all' });
    releaseLocalState();
  });
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
      send({ type: 'release_all' });
      releaseLocalState();
    }
  });
  if ('serviceWorker' in navigator && location.protocol === 'https:' && location.hostname !== 'touchbridge.appassets.local') navigator.serviceWorker.register('/sw.js').catch(() => {});
  rememberedServers = getRememberedServers();
  renderDeviceLists();
  if (location.hostname && location.hostname !== 'touchbridge.appassets.local' && location.protocol.startsWith('http')) {
    selectedServer = { name: 'This computer', url: location.origin, serverId: '', token: '' };
  }
  setStatus(false);
  setScreen('home');
  if (window.TouchBridgeAndroid?.scanServers) scanDevices();
})();
