/**
 * LightningRAG 主页 i18n：默认英文；按系统语言匹配，未匹配则英文；可本地持久化。
 * 依赖同目录 home-locales.js（ES module，本地打开 file:// 亦可加载）。
 */
import LOCALES from './home-locales.js';

const STORAGE_KEY = 'lr-home-locale';
const FALLBACK = 'en';

export const SUPPORTED_HOME_LOCALES = [
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

/** @param {string} raw BCP 47 语言标签 */
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
  for (const code of SUPPORTED_HOME_LOCALES) {
    if (code.toLowerCase() === l) return code;
  }
  const base = l.split('-')[0];
  for (const code of SUPPORTED_HOME_LOCALES) {
    if (code.split('-')[0].toLowerCase() === base) return code;
  }
  return null;
}

export function detectHomeLocale() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved && SUPPORTED_HOME_LOCALES.includes(saved)) return saved;
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
export function applyHomeLocale(locale) {
  if (!SUPPORTED_HOME_LOCALES.includes(locale)) locale = FALLBACK;
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

  const t = (k) => pathGet(dict, k);
  const title = t('meta.title');
  if (title) document.title = title;
  const mDesc = document.querySelector('meta[name="description"]');
  if (mDesc) {
    const v = t('meta.description');
    if (v) mDesc.setAttribute('content', v);
  }
  const mKw = document.querySelector('meta[name="keywords"]');
  if (mKw) {
    const v = t('meta.keywords');
    if (v) mKw.setAttribute('content', v);
  }
  const ogTitle = document.querySelector('meta[property="og:title"]');
  if (ogTitle) {
    const v = t('meta.ogTitle');
    if (v) ogTitle.setAttribute('content', v);
  }
  const ogDesc = document.querySelector('meta[property="og:description"]');
  if (ogDesc) {
    const v = t('meta.ogDescription');
    if (v) ogDesc.setAttribute('content', v);
  }

  const ld = document.querySelector('script[type="application/ld+json"]');
  if (ld) {
    const v = t('meta.jsonLdDescription');
    if (v) {
      try {
        const data = JSON.parse(ld.textContent);
        if (data && data['@type'] === 'SoftwareApplication') {
          data.description = v;
          ld.textContent = JSON.stringify(data);
        }
      } catch (_) {
        /* ignore */
      }
    }
  }

  const sel = document.getElementById('lr-lang-select');
  if (sel) sel.value = locale;
}

export function initHomeI18n() {
  const locale = detectHomeLocale();
  applyHomeLocale(locale);

  const sel = document.getElementById('lr-lang-select');
  if (sel) {
    sel.addEventListener('change', () => {
      const next = sel.value;
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch (_) {
        /* ignore */
      }
      applyHomeLocale(next);
    });
  }
}

initHomeI18n();
