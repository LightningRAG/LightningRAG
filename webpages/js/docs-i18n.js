/**
 * 文档站多语言：与主页共用 localStorage 键 lr-home-locale，语言选择器行为一致。
 */
import LOCALES from './docs-locales.js';

const STORAGE_KEY = 'lr-home-locale';
const FALLBACK = 'en';

export const SUPPORTED_DOC_LOCALES = [
  'en',
  'zh-CN',
  'zh-TW',
  'es',
  'fr',
  'de',
  'ja',
  'ko',
  'pt',
  'ru',
  'it',
];

function pathGet(obj, path) {
  return path.split('.').reduce((o, k) => {
    if (o == null || o[k] === undefined) return null;
    return o[k];
  }, obj);
}

export function matchSupportedLocale(raw) {
  if (!raw) return null;
  const l = String(raw).toLowerCase().replace(/_/g, '-');
  if (l === 'zh' || l.startsWith('zh-')) {
    if (
      l.includes('hant') ||
      l.endsWith('-tw') ||
      l.endsWith('-hk') ||
      l.endsWith('-mo')
    ) {
      return 'zh-TW';
    }
    return 'zh-CN';
  }
  if (l.startsWith('pt')) return 'pt';
  for (const code of SUPPORTED_DOC_LOCALES) {
    if (code.toLowerCase() === l) return code;
  }
  const base = l.split('-')[0];
  for (const code of SUPPORTED_DOC_LOCALES) {
    if (code.split('-')[0].toLowerCase() === base) return code;
  }
  return null;
}

export function detectDocLocale() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved && SUPPORTED_DOC_LOCALES.includes(saved)) return saved;
  } catch (_) {
    /* private mode */
  }
  const list =
    navigator.languages && navigator.languages.length
      ? navigator.languages
      : [navigator.language || 'en'];
  for (const raw of list) {
    const m = matchSupportedLocale(raw);
    if (m) return m;
  }
  return FALLBACK;
}

/**
 * @param {string} locale
 */
export function applyDocLocale(locale) {
  if (!SUPPORTED_DOC_LOCALES.includes(locale)) locale = FALLBACK;
  const dict = LOCALES[locale] || LOCALES[FALLBACK];
  document.documentElement.lang = locale;

  document.querySelectorAll('[data-i18n]').forEach((el) => {
    const key = el.getAttribute('data-i18n');
    const val = pathGet(dict, key);
    if (val != null) el.textContent = val;
  });
  document.querySelectorAll('[data-i18n-aria-label]').forEach((el) => {
    const key = el.getAttribute('data-i18n-aria-label');
    const val = pathGet(dict, key);
    if (val != null) el.setAttribute('aria-label', val);
  });
  document.querySelectorAll('[data-i18n-alt]').forEach((el) => {
    const key = el.getAttribute('data-i18n-alt');
    const val = pathGet(dict, key);
    if (val != null) el.setAttribute('alt', val);
  });
  document.querySelectorAll('[data-i18n-title]').forEach((el) => {
    const key = el.getAttribute('data-i18n-title');
    const val = pathGet(dict, key);
    if (val != null) el.setAttribute('title', val);
  });

  const t = (k) => pathGet(dict, k);
  const page = document.body?.getAttribute('data-doc-page') || 'hub';
  const metaPrefix =
    page === 'init'
      ? 'metaInit'
      : page === 'use'
        ? 'metaUse'
        : page === 'license'
          ? 'metaLicense'
          : page === 'preview'
            ? 'metaPreview'
            : 'metaHub';
  const title = t(`${metaPrefix}.title`);
  if (title) document.title = title;
  const mDesc = document.querySelector('meta[name="description"]');
  if (mDesc) {
    const v = t(`${metaPrefix}.description`);
    if (v) mDesc.setAttribute('content', v);
  }

  const sel = document.getElementById('lr-lang-select');
  if (sel) sel.value = locale;

  document.querySelectorAll('[data-i18n-locale-visible]').forEach((el) => {
    const raw = el.getAttribute('data-i18n-locale-visible') || '';
    const codes = raw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    if (codes.includes(locale)) el.removeAttribute('hidden');
    else el.setAttribute('hidden', '');
  });
}

export function initDocI18n() {
  const locale = detectDocLocale();
  applyDocLocale(locale);

  const sel = document.getElementById('lr-lang-select');
  if (sel) {
    sel.addEventListener('change', () => {
      const next = sel.value;
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch (_) {
        /* ignore */
      }
      applyDocLocale(next);
    });
  }
}

initDocI18n();
