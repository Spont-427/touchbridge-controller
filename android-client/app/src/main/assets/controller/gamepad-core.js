/* Pure controller math and built-in starter layouts. Kept dependency-free so
 * the same code can be exercised by Node tests and the Android WebView. */
(function (root) {
  'use strict';

  const BUTTONS = Object.freeze([
    'A', 'B', 'X', 'Y', 'LB', 'RB', 'BACK', 'START', 'GUIDE', 'LS', 'RS',
    'DPAD_UP', 'DPAD_DOWN', 'DPAD_LEFT', 'DPAD_RIGHT', 'LT', 'RT'
  ]);
  const clamp = (value, low, high) => Math.max(low, Math.min(high, value));
  const finite = value => Number.isFinite(value) ? value : 0;

  function normalizeStick(x, y, deadzone = 0.08, curve = 'linear') {
    x = finite(Number(x));
    y = finite(Number(y));
    deadzone = clamp(finite(Number(deadzone)), 0, 0.35);
    let magnitude = Math.hypot(x, y);
    if (magnitude < 1e-8 || magnitude <= deadzone) return { x: 0, y: 0, magnitude: 0 };

    const sourceMagnitude = Math.min(1, magnitude);
    const normalized = clamp((sourceMagnitude - deadzone) / (1 - deadzone), 0, 1);
    const exponent = curve === 'precision' ? 1.55 : curve === 'responsive' ? 0.78 : 1;
    const outputMagnitude = Math.pow(normalized, exponent);
    const scale = outputMagnitude / magnitude;
    const outX = clamp(x * scale, -1, 1);
    const outY = clamp(y * scale, -1, 1);
    return { x: outX, y: outY, magnitude: outputMagnitude };
  }

  // In floating/adaptive mode, the touch-down point becomes the temporary
  // origin. The origin is constrained to a configurable central activation
  // region so edge touches cannot create an unusably short stick throw.
  function adaptiveOrigin(x, y, width, height, activationRange = 0.48) {
    width = Math.max(1, finite(Number(width)));
    height = Math.max(1, finite(Number(height)));
    x = clamp(finite(Number(x)), 0, width);
    y = clamp(finite(Number(y)), 0, height);
    const cx = width / 2;
    const cy = height / 2;
    const radius = Math.min(width, height) / 2;
    const maxOffset = radius * clamp(finite(Number(activationRange)), 0, 0.8);
    const dx = x - cx;
    const dy = y - cy;
    const distance = Math.hypot(dx, dy);
    if (distance <= maxOffset || distance === 0) return { x, y };
    const scale = maxOffset / distance;
    return { x: cx + dx * scale, y: cy + dy * scale };
  }

  // Trigger controls are vertical pressure surfaces: top = fully pressed,
  // bottom = released. The value is clamped to [0, 1] and kept separate from
  // digital shoulder buttons.
  function triggerPressure(clientY, rectTop, height) {
    height = Math.max(1, finite(Number(height)));
    const y = finite(Number(clientY)) - finite(Number(rectTop));
    return clamp(1 - y / height, 0, 1);
  }

  // Pedals press downward: the bottom of the travel is full throttle/braking.
  function pedalPressure(clientY, rectTop, height) {
    return 1 - triggerPressure(clientY, rectTop, height);
  }

  // A steering wheel maps rotation around its centre onto the left stick's
  // horizontal axis. Positive rotation (clockwise) steers right.
  function wheelAxis(rotation, maxRotation = Math.PI * 0.75) {
    maxRotation = Math.max(0.01, finite(Number(maxRotation)));
    return clamp(finite(Number(rotation)) / maxRotation, -1, 1);
  }

  // Gyro steering is relative to a calibrated level position. A small deadzone
  // keeps hand tremor from nudging the car, and the result is clamped so a
  // steep tilt cannot exceed full lock.
  function steeringFromTilt(tilt, neutral = 0, maxTilt = 28, deadzone = 1.5) {
    const delta = finite(Number(tilt)) - finite(Number(neutral));
    maxTilt = Math.max(1, finite(Number(maxTilt)));
    deadzone = clamp(finite(Number(deadzone)), 0, maxTilt * 0.8);
    if (Math.abs(delta) <= deadzone) return 0;
    const direction = delta < 0 ? -1 : 1;
    return clamp((Math.abs(delta) - deadzone) / (maxTilt - deadzone) * direction, -1, 1);
  }

  // Do not steal keyboard navigation/activation from native controls. Using
  // closest() also handles events whose target is an icon/span inside a button.
  function isInteractiveTarget(target) {
    if (!target || typeof target.closest !== 'function') return false;
    return Boolean(target.closest('input, select, textarea, button, a, [contenteditable="true"], [role="button"]'));
  }

  const baseControls = [
    { id: 'left-stick', kind: 'stick', side: 'left', label: 'L', x: 23, y: 76, size: 136 },
    { id: 'right-stick', kind: 'stick', side: 'right', label: 'R', x: 77, y: 76, size: 136 },
    { id: 'dpad-up', kind: 'button', button: 'DPAD_UP', label: '▲', x: 22, y: 28, size: 43 },
    { id: 'dpad-left', kind: 'button', button: 'DPAD_LEFT', label: '◀', x: 14, y: 36, size: 43 },
    { id: 'dpad-right', kind: 'button', button: 'DPAD_RIGHT', label: '▶', x: 30, y: 36, size: 43 },
    { id: 'dpad-down', kind: 'button', button: 'DPAD_DOWN', label: '▼', x: 22, y: 44, size: 43 },
    { id: 'face-y', kind: 'button', button: 'Y', label: 'Y', x: 78, y: 24, size: 46, hint: 'Top face button' },
    { id: 'face-x', kind: 'button', button: 'X', label: 'X', x: 70, y: 32, size: 46 },
    { id: 'face-b', kind: 'button', button: 'B', label: 'B', x: 86, y: 32, size: 46 },
    { id: 'face-a', kind: 'button', button: 'A', label: 'A', x: 78, y: 40, size: 46 },
    { id: 'left-bumper', kind: 'button', button: 'LB', label: 'LB', x: 25, y: 9, size: 48 },
    { id: 'left-trigger', kind: 'button', button: 'LT', label: 'LT', x: 12, y: 9, size: 48 },
    { id: 'right-bumper', kind: 'button', button: 'RB', label: 'RB', x: 75, y: 9, size: 48 },
    { id: 'right-trigger', kind: 'button', button: 'RT', label: 'RT', x: 88, y: 9, size: 48 },
    { id: 'back', kind: 'button', button: 'BACK', label: 'BACK', x: 43, y: 55, size: 48 },
    { id: 'start', kind: 'button', button: 'START', label: 'START', x: 57, y: 55, size: 48 }
  ];

  const PROFILE_META = Object.freeze({
    generic: {
      name: 'Generic controller',
      description: 'Balanced two-stick layout for games that use a standard Xbox-style controller.'
    },
    gta: {
      name: 'GTA V',
      description: 'Starter layout for driving, aiming and on-foot movement. Verify and customize controls for your in-game preset.'
    },
    rdr2: {
      name: 'Red Dead Redemption 2',
      description: 'Starter layout for movement, aiming and interaction. Game and platform control schemes can vary.'
    },
    forza: {
      name: 'Forza Horizon',
      description: 'Driving-oriented layout with larger trigger controls for acceleration and braking.'
    }
  });

  function createProfile(profileId) {
    const controls = baseControls.map(control => ({ ...control }));
    if (profileId === 'gta') {
      Object.assign(controls.find(c => c.id === 'left-stick'), { x: 22, y: 73, size: 142 });
      Object.assign(controls.find(c => c.id === 'right-stick'), { x: 78, y: 73, size: 142 });
      Object.assign(controls.find(c => c.id === 'dpad-up'), { x: 19, y: 26 });
      Object.assign(controls.find(c => c.id === 'dpad-down'), { x: 19, y: 44 });
      Object.assign(controls.find(c => c.id === 'left-trigger'), { x: 11, y: 9, size: 54 });
      Object.assign(controls.find(c => c.id === 'right-trigger'), { x: 89, y: 9, size: 54 });
    } else if (profileId === 'rdr2') {
      Object.assign(controls.find(c => c.id === 'left-stick'), { x: 21, y: 75, size: 140 });
      Object.assign(controls.find(c => c.id === 'right-stick'), { x: 79, y: 75, size: 140 });
      Object.assign(controls.find(c => c.id === 'face-a'), { x: 78, y: 42, size: 49 });
      Object.assign(controls.find(c => c.id === 'face-b'), { x: 87, y: 33, size: 49 });
    } else if (profileId === 'forza') {
      Object.assign(controls.find(c => c.id === 'left-stick'), { x: 22, y: 76, size: 126 });
      Object.assign(controls.find(c => c.id === 'right-stick'), { x: 77, y: 76, size: 126 });
      Object.assign(controls.find(c => c.id === 'left-trigger'), { x: 11, y: 10, size: 61 });
      Object.assign(controls.find(c => c.id === 'right-trigger'), { x: 89, y: 10, size: 61 });
      Object.assign(controls.find(c => c.id === 'left-bumper'), { x: 27, y: 10, size: 45 });
      Object.assign(controls.find(c => c.id === 'right-bumper'), { x: 73, y: 10, size: 45 });
    }
    return controls;
  }



  function snapPercent(value, enabled = true, step = 2) {
    value = finite(Number(value));
    step = clamp(finite(Number(step)), 1, 20);
    return clamp(enabled ? Math.round(value / step) * step : value, 4, 96);
  }

  function findOpenPosition(layout, size = 46, stageWidth = 0, stageHeight = 0) {
    const candidates = [
      { x: 50, y: 16 }, { x: 50, y: 84 }, { x: 40, y: 18 }, { x: 60, y: 18 },
      { x: 40, y: 82 }, { x: 60, y: 82 }, { x: 50, y: 32 }, { x: 50, y: 68 },
      { x: 36, y: 50 }, { x: 64, y: 50 }, { x: 34, y: 18 }, { x: 66, y: 18 },
      { x: 34, y: 82 }, { x: 66, y: 82 }, { x: 50, y: 50 }
    ];
    const existing = Array.isArray(layout) ? layout : [];
    const newSize = Math.max(1, Number(size) || 46);
    const width = Number(stageWidth) > 0 && Number(stageHeight) > 0 ? Number(stageWidth) : 0;
    const height = Number(stageHeight) > 0 ? Number(stageHeight) : 0;
    const halfW = width ? (newSize / 2 / width) * 100 : 0;
    const halfH = height ? (newSize / 2 / height) * 100 : 0;
    let best = candidates[0];
    let bestScore = -Infinity;
    for (const candidate of candidates) {
      if (width && height && (
        candidate.x - halfW < 4 || candidate.x + halfW > 96 ||
        candidate.y - halfH < 4 || candidate.y + halfH > 96
      )) continue;
      const nearest = existing.reduce((minimum, item) => {
        if (!item || !Number.isFinite(item.x) || !Number.isFinite(item.y)) return minimum;
        if (width && height) {
          const otherSize = Number(item.size) || 46;
          const gapX = Math.abs(candidate.x - item.x) - halfW - (otherSize / 2 / width) * 100;
          const gapY = Math.abs(candidate.y - item.y) - halfH - (otherSize / 2 / height) * 100;
          // A full 2D overlap is much worse than a near miss, so rank it below
          // every candidate that has at least one axis clear.
          const clearance = Math.max(gapX, gapY);
          return Math.min(minimum, clearance);
        }
        const dx = (candidate.x - item.x) * 0.62;
        const dy = candidate.y - item.y;
        const clearance = Math.hypot(dx, dy) - Math.max(5, (Number(item.size) || 46) * 0.11);
        return Math.min(minimum, clearance);
      }, Infinity);
      if (nearest > bestScore) { bestScore = nearest; best = candidate; }
    }
    return { x: best.x, y: best.y };
  }

  function validateLayout(layout) {
    if (!Array.isArray(layout) || layout.length < 2 || layout.length > 40) return false;
    const ids = new Set();
    const stickSides = new Set();
    let wheels = 0;
    for (const item of layout) {
      if (!item || typeof item.id !== 'string' || !/^[a-zA-Z0-9_-]{1,64}$/.test(item.id) || ids.has(item.id)) return false;
      ids.add(item.id);
      if (!['stick', 'button', 'wheel'].includes(item.kind)) return false;
      if (item.kind === 'stick') {
        if (!['left', 'right'].includes(item.side) || stickSides.has(item.side)) return false;
        stickSides.add(item.side);
      }
      if (item.kind === 'button' && !BUTTONS.includes(item.button)) return false;
      if (item.kind === 'wheel' && ++wheels > 1) return false;
      if (item.pedal !== undefined && typeof item.pedal !== 'boolean') return false;
      if (![item.x, item.y, item.size].every(Number.isFinite)) return false;
      if (item.x < 3 || item.x > 97 || item.y < 3 || item.y > 97 || item.size < 32 || item.size > 190) return false;
      if (typeof item.label !== 'string' || item.label.length > 12) return false;
      if (item.opacity !== undefined && (!Number.isFinite(item.opacity) || item.opacity < 0.25 || item.opacity > 1)) return false;
      if (item.hint !== undefined && (typeof item.hint !== 'string' || item.hint.length > 100)) return false;
    }
    return stickSides.has('left') && stickSides.has('right');
  }

  const api = { BUTTONS, clamp, normalizeStick, adaptiveOrigin, triggerPressure, pedalPressure, wheelAxis, steeringFromTilt, snapPercent, findOpenPosition, isInteractiveTarget, PROFILE_META, createProfile, validateLayout };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  root.TouchBridgeGamepad = api;
})(typeof globalThis !== 'undefined' ? globalThis : this);
