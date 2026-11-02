import assert from 'node:assert/strict';
import test from 'node:test';
import { checkedReleaseVersion } from './check_release_version.mjs';

test('tag and npm package use one version', () => {
  assert.equal(checkedReleaseVersion('v1.2.3', '1.2.3'), '1.2.3');
  assert.equal(checkedReleaseVersion('1.2.3-rc.1', '1.2.3-rc.1'), '1.2.3-rc.1');
});
test('different asset and npm versions fail before publication', () => {
  assert.throws(() => checkedReleaseVersion('1.2.4', '1.2.3'), /does not match/);
});
test('missing version cannot create a release', () => {
  assert.throws(() => checkedReleaseVersion(undefined, '1.2.3'), /missing/);
});
