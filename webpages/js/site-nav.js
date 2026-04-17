/**
 * 窄屏顶栏：折叠主导航，按钮展开/收起；宽屏由 CSS 隐藏按钮并保持横向导航。
 */
(function () {
  const MQ = window.matchMedia('(min-width: 901px)');

  function navElements() {
    const btn = document.querySelector('.site-nav-toggle');
    const nav = document.getElementById('lr-site-nav');
    return { btn, nav };
  }

  function setOpen(open) {
    const { btn, nav } = navElements();
    if (!btn || !nav) return;
    document.body.classList.toggle('lr-nav-open', open);
    btn.setAttribute('aria-expanded', open ? 'true' : 'false');
  }

  function closeIfDesktop() {
    if (MQ.matches) setOpen(false);
  }

  function init() {
    const { btn, nav } = navElements();
    if (!btn || !nav) return;

    btn.addEventListener('click', () => {
      const open = !document.body.classList.contains('lr-nav-open');
      setOpen(open);
    });

    nav.querySelectorAll('a').forEach((a) => {
      a.addEventListener('click', () => setOpen(false));
    });

    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') setOpen(false);
    });

    document.addEventListener('click', (e) => {
      if (!document.body.classList.contains('lr-nav-open')) return;
      const t = e.target;
      if (t instanceof Node && !nav.contains(t) && !btn.contains(t)) setOpen(false);
    });

    MQ.addEventListener('change', closeIfDesktop);
    window.addEventListener('resize', closeIfDesktop);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
