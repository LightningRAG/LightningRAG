const test = require('node:test');
const assert = require('node:assert/strict');
const crypto = require('crypto');
const fs = require('fs');
const os = require('os');
const path = require('path');
const vm = require('vm');

const { svgBuilder } = require('..');

const AUTH_KEY = '04788f1ea15d305f';

function buildAuthorizedSecret(projectName) {
  const domain = new URL(projectName).hostname.replace(/^www\./, '');
  const projectNameMd5 = crypto.createHash('md5').update(domain).digest('hex');
  const maxLength = projectNameMd5.length;

  let sec = '';
  for (let index = 0; index < maxLength; index += 1) {
    sec += projectNameMd5[index];
    sec += AUTH_KEY[index] || 'x';
    sec += 'y';
  }

  return sec;
}

test('authorized production build injects parseable domain verification script', () => {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'vite-auto-import-svg-'));
  const iconsDir = path.join(tempRoot, 'icons');
  fs.mkdirSync(iconsDir);
  fs.writeFileSync(
    path.join(iconsDir, 'logo.svg'),
    '<svg width="24" height="24"><path d="M0 0h24v24H0z" /></svg>'
  );

  const projectName = 'https://example.com';
  global['gva-project-name'] = projectName;
  global['gva-secret'] = buildAuthorizedSecret(projectName);

  const plugin = svgBuilder([`${iconsDir.replace(/\\/g, '/')}/`], '/', 'dist', 'assets', 'production');
  const bundle = {
    'assets/index.js': {
      type: 'chunk',
      isEntry: true,
      code: 'console.log("entry");'
    }
  };

  plugin.generateBundle({}, bundle);
  assert.match(bundle['assets/index.js'].code, /var _expected=/);

  assert.doesNotThrow(() => {
    new vm.Script(bundle['assets/index.js'].code);
  });
});
