const test = require('node:test');
const assert = require('node:assert/strict');
const core = require('./gamepad-core.js');

test('radial deadzone zeros small inputs and preserves a full axis', () => {
  assert.deepEqual(core.normalizeStick(0.02, 0.02, 0.1, 'linear'), { x: 0, y: 0, magnitude: 0 });
  const full = core.normalizeStick(1, 0, 0.1, 'linear');
  assert.equal(full.x, 1);
  assert.equal(full.y, 0);
});

test('radial normalization caps diagonal magnitude and supports curves', () => {
  const diagonal = core.normalizeStick(1, 1, 0, 'linear');
  assert.ok(diagonal.x <= 1 && diagonal.y <= 1);
  assert.ok(Math.hypot(diagonal.x, diagonal.y) <= 1.000001);
  const precision = core.normalizeStick(0.5, 0, 0.05, 'precision');
  const responsive = core.normalizeStick(0.5, 0, 0.05, 'responsive');
  assert.ok(precision.x < responsive.x);
});

test('stick math rejects non-finite coordinates safely', () => {
  assert.deepEqual(core.normalizeStick(Number.NaN, Number.POSITIVE_INFINITY), { x: 0, y: 0, magnitude: 0 });
});

test('adaptive origin uses the tap point inside the activation region', () => {
  const origin = core.adaptiveOrigin(70, 70, 200, 200, 0.5);
  assert.deepEqual(origin, { x: 70, y: 70 });
});

test('adaptive origin clamps distant edge taps to the activation boundary', () => {
  const origin = core.adaptiveOrigin(200, 100, 200, 200, 0.4);
  assert.equal(origin.x, 140);
  assert.equal(origin.y, 100);
});

test('built-in profile layouts are valid and contain two sticks', () => {
  for (const id of Object.keys(core.PROFILE_META)) {
    const layout = core.createProfile(id);
    assert.ok(core.validateLayout(layout), `invalid layout: ${id}`);
    assert.equal(layout.filter(c => c.kind === 'stick').length, 2);
  }
});


test('analog trigger pressure maps the top to full and bottom to released', () => {
  assert.equal(core.triggerPressure(100, 100, 50), 1);
  assert.equal(core.triggerPressure(125, 100, 50), 0.5);
  assert.equal(core.triggerPressure(150, 100, 50), 0);
  assert.equal(core.triggerPressure(999, 100, 50), 0);
});

test('layout validation requires one left and one right stick', () => {
  const valid = core.createProfile('generic');
  assert.equal(core.validateLayout(valid), true);
  assert.equal(core.validateLayout(valid.filter(item => item.side !== 'left')), false);
  const duplicate = valid.map(item => ({ ...item }));
  duplicate.find(item => item.id === 'right-stick').side = 'left';
  assert.equal(core.validateLayout(duplicate), false);
});

test('layout validation rejects unsafe identifiers and unsupported actions', () => {
  const unsafeId = core.createProfile('generic');
  unsafeId[0].id = '<script>';
  assert.equal(core.validateLayout(unsafeId), false);
  const badAction = core.createProfile('generic');
  badAction.find(item => item.kind === 'button').button = 'RUN_CODE';
  assert.equal(core.validateLayout(badAction), false);
});


test('keyboard forwarding leaves native interactive controls alone', () => {
  const selector = 'input, select, textarea, button, a, [contenteditable="true"], [role="button"]';
  assert.equal(core.isInteractiveTarget({ closest: value => value === selector ? { tagName: 'BUTTON' } : null }), true);
  assert.equal(core.isInteractiveTarget({ closest: value => value === selector ? null : null }), false);
  assert.equal(core.isInteractiveTarget(null), false);
});


test('layout opacity is optional for old profiles and bounded for edited layouts', () => {
  const legacy = core.createProfile('generic');
  assert.equal(core.validateLayout(legacy), true);
  const edited = legacy.map(item => ({ ...item, opacity: 0.6 }));
  assert.equal(core.validateLayout(edited), true);
  edited[0].opacity = 0.1;
  assert.equal(core.validateLayout(edited), false);
  edited[0].opacity = 1.2;
  assert.equal(core.validateLayout(edited), false);
});


test('new controls are placed away from the existing layout where practical', () => {
  const existing = core.createProfile('generic');
  const position = core.findOpenPosition(existing);
  assert.ok(position.x >= 3 && position.x <= 97);
  assert.ok(position.y >= 3 && position.y <= 97);
  assert.equal(existing.some(item => item.x === position.x && item.y === position.y), false);
  assert.deepEqual(core.findOpenPosition(existing), position);
});


test('editor snap-to-grid is optional and clamps controls inside the canvas', () => {
  assert.equal(core.snapPercent(13.2, true, 2), 14);
  assert.equal(core.snapPercent(13.2, false, 2), 13.2);
  assert.equal(core.snapPercent(-30, true, 2), 4);
  assert.equal(core.snapPercent(120, true, 2), 96);
});
