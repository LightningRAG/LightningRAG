/**
 * 主页文案包：多语言（en 默认 + 含简繁中文与主流语言）
 * 键与 index.html 中 data-i18n 路径一致（点分嵌套）
 */
export default {
  en: {
    meta: {
      title:
        'LightningRAG — High-performance full-stack RAG and development platform',
      description:
        'LightningRAG is a Go (Gin) and Vue 3 full-stack RAG and admin foundation. Compared with typical Python stacks, it offers higher performance, compiled deployment, smaller artifacts, and stronger concurrency.',
      keywords:
        'LightningRAG,RAG,retrieval-augmented generation,Gin,Vue 3,Go,full-stack,admin,knowledge base,open source',
      ogTitle:
        'LightningRAG — High-performance full-stack RAG and development platform',
      ogDescription:
        'Enterprise RAG and admin platform built with Go and Vue. Fast, shippable as binaries, compact, and scalable concurrency.',
      jsonLdDescription:
        'Full-stack RAG and admin foundation with Gin and Vue 3, featuring JWT, dynamic routes, permissions, and code generation.',
    },
    a11y: {
      skipMain: 'Skip to main content',
      navMain: 'Main navigation',
      brandHome: 'LightningRAG home',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Features',
      advantages: 'Advantages',
      articles: 'Articles',
      docs: 'Documentation',
      license: 'Commercial license',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Full-stack RAG platform',
      title: 'A high-performance foundation for enterprise RAG and business backends',
      sub:
        'LightningRAG brings together authentication, dynamic routing, knowledge bases, and agents so you can focus on business logic and model quality—not rebuilding infrastructure.',
      ctaSource: 'View source on GitHub',
      ctaDocs: 'Read the docs',
    },
    features: {
      title: 'Product highlights',
      lead:
        'Extend a proven full-stack architecture for RAG—from knowledge bases to agent orchestration—with one unified stack and delivery model.',
      c1t: 'Full-stack, unified',
      c1p:
        'Gin on the backend and Vue 3 on the frontend, plus JWT, dynamic menus, Casbin, and code generation—ship admin consoles and RAG pipelines faster.',
      c2t: 'RAG-ready workflows',
      c2p:
        'Retrieval-augmented pipelines where knowledge, documents, and agents integrate with your users and permissions—ideal for private deployments and customization.',
      c3t: 'Extensible and integrable',
      c3p:
        'Modular hooks for vector stores, model services, and enterprise SSO. Website content and documentation can evolve together with articles and release notes.',
    },
    adv: {
      title: 'Core advantages vs. typical Python projects',
      lead:
        'A Go-based compiled backend complements interpreted Python stacks in performance, shipping model, and operations—especially when latency, concurrency, and deployment footprint matter.',
      s1t: 'Speed and throughput',
      s1p:
        'Native binaries start quickly with predictable GC. For high-concurrency retrieval and API aggregation, you typically get steadier latency and higher throughput—and simpler horizontal scaling.',
      s2t: 'Compiled delivery and stronger code protection',
      s2p:
        'Ship one or a few binaries without exposing a plain source tree on the server—often better for protecting implementation details than typical Python source deployments.',
      s3t: 'Smaller deployment footprint',
      s3p:
        'No need to ship a full interpreter and large runtime stacks for the core service—leaner images and artifacts, including at the edge and in resource-constrained environments.',
      s4t: 'Higher concurrent user capacity',
      s4p:
        'Goroutines and lower baseline memory suit long-lived connections and multi-tenant scenarios—often more simultaneous sessions and admin operations on the same hardware.',
    },
    bottom: {
      title: 'Try it and contribute',
      lead: 'Licensed under Apache 2.0. Stars, issues, and pull requests are welcome.',
      commercial: 'Need a commercial license? Contact us.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: 'Browse articles (updated over time)',
    },
    footer: {
      copy: '© LightningRAG open-source community. All rights reserved.',
      docs: 'Documentation',
      githubRepo: 'GitHub repository',
      articles: 'Articles',
      license: 'Commercial license',
      sitemap: 'Sitemap',
    },
    ui: { langAria: 'Language' },
  },
  'zh-CN': {
    meta: {
      title: 'LightningRAG — 高性能全栈 RAG 与开发基础平台',
      description:
        'LightningRAG — 基于 Go（Gin）与 Vue 3 的全栈 RAG 与后台基础平台。相较典型 Python 方案，具备更高性能、编译部署、更小体积与更强并发承载能力。',
      keywords:
        'LightningRAG,RAG,检索增强生成,Gin,Vue3,Go,全栈框架,知识库,智能问答,开源',
      ogTitle: 'LightningRAG — 高性能全栈 RAG 与开发基础平台',
      ogDescription:
        'Go + Vue 构建的企业级 RAG 与管理后台基础平台。速度快、可编译交付、体积小、易扩展并发。',
      jsonLdDescription:
        '基于 Go（Gin）与 Vue 3 的全栈 RAG 与后台基础平台，支持 JWT、动态路由、权限与代码生成等能力。',
    },
    a11y: {
      skipMain: '跳到主要内容',
      navMain: '主导航',
      brandHome: 'LightningRAG 首页',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: '产品特点',
      advantages: '核心优势',
      articles: '文章',
      docs: '文档',
      license: '商用授权',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · 全栈 RAG 基础平台',
      title: '为企业级 RAG 与业务后台而生的高性能基座',
      sub:
        'LightningRAG 整合鉴权、动态路由、知识库与智能体等能力，让你把精力集中在业务与模型效果上，而不是重复搭建基础设施。',
      ctaSource: '在 GitHub 上查看源码',
      ctaDocs: '阅读文档',
    },
    features: {
      title: '产品特点',
      lead:
        '在成熟的全栈架构上扩展 RAG 场景：从知识库管理到智能体编排，与后台体系统一技术栈、统一交付方式。',
      c1t: '全栈一体化',
      c1p:
        '后端 Gin、前端 Vue 3，前后端分离；内置 JWT、动态菜单、Casbin 权限、代码生成器等，快速落地管理端与 RAG 管线。',
      c2t: 'RAG 场景就绪',
      c2p:
        '面向检索增强生成的工作流：知识库、文档与智能体能力可与现有权限、用户体系无缝衔接，适合私有化与二次开发。',
      c3t: '可扩展、可集成',
      c3p:
        '模块化设计便于接入向量库、模型服务与企业 SSO；官网内容与文档可并行演进，方便持续发布文章与更新日志。',
    },
    adv: {
      title: '相较典型 Python 类项目的核心优势',
      lead:
        '后端采用 Go 编译型语言构建服务，在性能、交付形态与运维成本上与解释型 Python 栈形成互补，尤其适合对延迟、并发与部署形态有要求的场景。',
      s1t: '运行速度与吞吐',
      s1p:
        '编译为原生机器码，启动快、GC 可控，在高并发检索与 API 聚合场景下通常具备更稳定的延迟与更高吞吐，便于横向扩展。',
      s2t: '编译交付与代码保护',
      s2p:
        '发布产物为单一或少量二进制文件，不含明文业务源码目录；相对依赖源码部署的 Python 项目，更有利于保护实现细节与内部逻辑。',
      s3t: '部署体积小',
      s3p:
        '无需在服务器携带完整解释器与大量运行时依赖即可运行核心服务；镜像与制品更精简，利于边缘节点与资源受限环境。',
      s4t: '支撑更多并发用户',
      s4p:
        '协程模型与较低内存占用适合长连接与多租户场景；在同等硬件下往往可承载更多同时在线会话与后台管理操作。',
    },
    bottom: {
      title: '开始试用与共建',
      lead: '开源协议 Apache 2.0。欢迎 Star、Issue 反馈与 PR 贡献。',
      commercial: '需要商用授权？联系我们。',
      github: 'GitHub：LightningRAG/LightningRAG',
      articles: '浏览文章（持续更新）',
    },
    footer: {
      copy: '© LightningRAG 开源社区。保留所有权利。',
      docs: '文档',
      githubRepo: 'GitHub 仓库',
      articles: '文章',
      license: '商用授权',
      sitemap: '站点地图',
    },
    ui: { langAria: '语言' },
  },
  'zh-TW': {
    meta: {
      title: 'LightningRAG — 高效能全端 RAG 與開發基礎平台',
      description:
        'LightningRAG — 以 Go（Gin）與 Vue 3 打造的全端 RAG 與後台基礎平台。相較典型 Python 方案，具備更高效能、編譯部署、更小體積與更強並行承載能力。',
      keywords:
        'LightningRAG,RAG,檢索增強生成,Gin,Vue3,Go,全端框架,知識庫,智慧問答,開源',
      ogTitle: 'LightningRAG — 高效能全端 RAG 與開發基礎平台',
      ogDescription:
        '以 Go 與 Vue 建置的企業級 RAG 與管理後台基礎平台。速度快、可編譯交付、體積小、易擴展並行處理。',
      jsonLdDescription:
        '以 Go（Gin）與 Vue 3 打造的全端 RAG 與後台基礎平台，支援 JWT、動態路由、權限與程式碼產生等能力。',
    },
    a11y: {
      skipMain: '跳到主要內容',
      navMain: '主要導覽',
      brandHome: 'LightningRAG 首頁',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: '產品特點',
      advantages: '核心優勢',
      articles: '文章',
      docs: '文件',
      license: '商用授權',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · 全端 RAG 基礎平台',
      title: '為企業級 RAG 與業務後台而生的高效能基座',
      sub:
        'LightningRAG 整合驗證與授權、動態路由、知識庫與智慧體等能力，讓您把心力放在業務與模型成效上，而不是重複搭建基礎設施。',
      ctaSource: '在 GitHub 上查看原始碼',
      ctaDocs: '閱讀文件',
    },
    features: {
      title: '產品特點',
      lead:
        '在成熟的全端架構上擴展 RAG 情境：從知識庫管理到智慧體編排，與後台體系統一技術棧、統一交付方式。',
      c1t: '全端一體化',
      c1p:
        '後端 Gin、前端 Vue 3，前後端分離；內建 JWT、動態選單、Casbin 權限、程式碼產生器等，快速上線管理端與 RAG 管線。',
      c2t: 'RAG 情境就緒',
      c2p:
        '面向檢索增強生成的工作流：知識庫、文件與智慧體能力可與現有權限、使用者體系無縫銜接，適合私有化與二次開發。',
      c3t: '可擴展、可整合',
      c3p:
        '模組化設計便於接入向量資料庫、模型服務與企業 SSO；官網內容與文件可並行演進，方便持續發布文章與更新日誌。',
    },
    adv: {
      title: '相較典型 Python 類專案的核心優勢',
      lead:
        '後端採用 Go 編譯型語言建構服務，在效能、交付型態與維運成本上與直譯型 Python 棧形成互補，尤其適合對延遲、並行與部署型態有要求的場景。',
      s1t: '執行速度與吞吐',
      s1p:
        '編譯為原生機器碼，啟動快、GC 可控，在高並行檢索與 API 聚合場景下通常具備更穩定的延遲與更高吞吐，便於橫向擴展。',
      s2t: '編譯交付與程式碼保護',
      s2p:
        '發布產物為單一或少量二進位檔，不含明文業務原始碼目錄；相對依賴原始碼部署的 Python 專案，更有利於保護實作細節與內部邏輯。',
      s3t: '部署體積小',
      s3p:
        '無需在伺服器攜帶完整直譯器與大量執行時依賴即可運行核心服務；映像與產物更精簡，利於邊緣節點與資源受限環境。',
      s4t: '支撐更多並行使用者',
      s4p:
        '協程模型與較低記憶體占用適合長連線與多租戶場景；在同等硬體下往往可承載更多同時上線工作階段與後台管理操作。',
    },
    bottom: {
      title: '開始試用與共建',
      lead: '開源授權 Apache 2.0。歡迎 Star、Issue 回報與 PR 貢獻。',
      commercial: '需要商用授權？與我們聯絡。',
      github: 'GitHub：LightningRAG/LightningRAG',
      articles: '瀏覽文章（持續更新）',
    },
    footer: {
      copy: '© LightningRAG 開源社群。保留所有權利。',
      docs: '文件',
      githubRepo: 'GitHub 儲存庫',
      articles: '文章',
      license: '商用授權',
      sitemap: '網站地圖',
    },
    ui: { langAria: '語言' },
  },
  es: {
    meta: {
      title:
        'LightningRAG — Plataforma RAG y de desarrollo full stack de alto rendimiento',
      description:
        'LightningRAG es una base RAG y de administración full stack con Go (Gin) y Vue 3. Frente a pilas Python típicas ofrece mayor rendimiento, despliegue compilado, artefactos más pequeños y mayor concurrencia.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,full stack,admin,base de conocimiento,código abierto',
      ogTitle:
        'LightningRAG — Plataforma RAG y de desarrollo full stack de alto rendimiento',
      ogDescription:
        'Plataforma empresarial RAG y de administración con Go y Vue. Rápida, entregable como binarios, compacta y con concurrencia escalable.',
      jsonLdDescription:
        'Base RAG y de administración full stack con Gin y Vue 3: JWT, rutas dinámicas, permisos y generación de código.',
    },
    a11y: {
      skipMain: 'Ir al contenido principal',
      navMain: 'Navegación principal',
      brandHome: 'Inicio de LightningRAG',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Características',
      advantages: 'Ventajas',
      articles: 'Artículos',
      docs: 'Documentación',
      license: 'Licencia comercial',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Plataforma RAG full stack',
      title:
        'Una base de alto rendimiento para RAG empresarial y backends de negocio',
      sub:
        'LightningRAG reúne autenticación, enrutamiento dinámico, bases de conocimiento y agentes para que te centres en la lógica de negocio y la calidad del modelo, no en rehacer infraestructura.',
      ctaSource: 'Ver código en GitHub',
      ctaDocs: 'Leer la documentación',
    },
    features: {
      title: 'Aspectos destacados',
      lead:
        'Extiende una arquitectura full stack probada para RAG—desde bases de conocimiento hasta orquestación de agentes—con una sola pila y un solo modelo de entrega.',
      c1t: 'Full stack unificado',
      c1p:
        'Gin en el backend y Vue 3 en el frontend, más JWT, menús dinámicos, Casbin y generación de código: administra consolas y tuberías RAG más rápido.',
      c2t: 'Flujos listos para RAG',
      c2p:
        'Tuberías de recuperación aumentada donde conocimiento, documentos y agentes se integran con usuarios y permisos—ideal para despliegues privados y personalización.',
      c3t: 'Extensible e integrable',
      c3p:
        'Ganchos modulares para almacenes vectoriales, servicios de modelos y SSO empresarial. El sitio y la documentación pueden evolucionar junto con artículos y notas de versión.',
    },
    adv: {
      title: 'Ventajas frente a proyectos Python típicos',
      lead:
        'Un backend compilado en Go complementa pilas Python interpretadas en rendimiento, forma de entrega y operaciones—especialmente cuando importan latencia, concurrencia y tamaño del despliegue.',
      s1t: 'Velocidad y rendimiento',
      s1p:
        'Los binarios nativos arrancan rápido con GC predecible. En recuperación concurrente y agregación de API sueles obtener latencia más estable y mayor rendimiento—y escalar horizontalmente es más sencillo.',
      s2t: 'Entrega compilada y mejor protección del código',
      s2p:
        'Entrega uno o pocos binarios sin exponer el árbol de fuentes en texto plano en el servidor—a menudo mejor para proteger detalles de implementación que despliegues Python típicos desde código.',
      s3t: 'Huella de despliegue más pequeña',
      s3p:
        'No necesitas enviar un intérprete completo y grandes pilas de runtime para el servicio principal—imágenes y artefactos más livianos, también en el borde y entornos con recursos limitados.',
      s4t: 'Mayor capacidad de usuarios concurrentes',
      s4p:
        'Las goroutines y un uso base de memoria menor encajan con conexiones largas y escenarios multiinquilino—a menudo más sesiones simultáneas y operaciones de administración con el mismo hardware.',
    },
    bottom: {
      title: 'Pruébalo y contribuye',
      lead: 'Licencia Apache 2.0. Stars, incidencias y pull requests son bienvenidos.',
      commercial: '¿Necesitas una licencia comercial? Contáctanos.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: 'Ver artículos (actualizados con el tiempo)',
    },
    footer: {
      copy: '© Comunidad de código abierto LightningRAG. Todos los derechos reservados.',
      docs: 'Documentación',
      githubRepo: 'Repositorio en GitHub',
      articles: 'Artículos',
      license: 'Licencia comercial',
      sitemap: 'Mapa del sitio',
    },
    ui: { langAria: 'Idioma' },
  },
  fr: {
    meta: {
      title:
        'LightningRAG — Plateforme RAG full stack haute performance pour le développement',
      description:
        'LightningRAG est une base RAG et d’administration full stack Go (Gin) et Vue 3. Par rapport aux piles Python classiques, elle offre de meilleures performances, un déploiement compilé, des artefacts plus compacts et une concurrence plus forte.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,full stack,admin,base de connaissances,open source',
      ogTitle:
        'LightningRAG — Plateforme RAG full stack haute performance pour le développement',
      ogDescription:
        'Plateforme RAG et d’administration d’entreprise avec Go et Vue. Rapide, livrable en binaires, compacte et à concurrence scalable.',
      jsonLdDescription:
        'Base RAG et d’administration full stack avec Gin et Vue 3 : JWT, routes dynamiques, droits et génération de code.',
    },
    a11y: {
      skipMain: 'Aller au contenu principal',
      navMain: 'Navigation principale',
      brandHome: 'Accueil LightningRAG',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Fonctionnalités',
      advantages: 'Avantages',
      articles: 'Articles',
      docs: 'Documentation',
      license: 'Licence commerciale',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Plateforme RAG full stack',
      title:
        'Une base haute performance pour le RAG d’entreprise et les backends métier',
      sub:
        'LightningRAG réunit authentification, routage dynamique, bases de connaissances et agents pour vous concentrer sur la logique métier et la qualité des modèles—pas sur la reconstruction d’infrastructure.',
      ctaSource: 'Voir le code sur GitHub',
      ctaDocs: 'Lire la documentation',
    },
    features: {
      title: 'Points forts',
      lead:
        'Étendez une architecture full stack éprouvée pour le RAG—des bases de connaissances à l’orchestration d’agents—avec une pile et un modèle de livraison unifiés.',
      c1t: 'Full stack unifié',
      c1p:
        'Gin côté serveur et Vue 3 côté client, avec JWT, menus dynamiques, Casbin et génération de code : déployez plus vite consoles d’admin et pipelines RAG.',
      c2t: 'Flux prêts pour le RAG',
      c2p:
        'Pipelines d’amplification par la récupération où connaissances, documents et agents s’intègrent à vos utilisateurs et permissions—idéal pour déploiements privés et personnalisation.',
      c3t: 'Extensible et intégrable',
      c3p:
        'Points d’extension pour magasins vectoriels, services de modèles et SSO d’entreprise. Le site et la documentation évoluent avec articles et notes de version.',
    },
    adv: {
      title: 'Avantages par rapport aux projets Python typiques',
      lead:
        'Un backend compilé en Go complète les piles Python interprétées sur le plan des performances, du mode de livraison et des opérations—surtout quand latence, concurrence et empreinte de déploiement comptent.',
      s1t: 'Vitesse et débit',
      s1p:
        'Les binaires natifs démarrent vite avec un GC prévisible. Pour la récupération à forte concurrence et l’agrégation d’API, vous obtenez souvent une latence plus stable et un débit plus élevé—avec un scale-out plus simple.',
      s2t: 'Livraison compilée et meilleure protection du code',
      s2p:
        'Livrez un ou quelques binaires sans exposer l’arborescence source en clair sur le serveur—souvent mieux pour protéger les détails d’implémentation que les déploiements Python depuis les sources.',
      s3t: 'Empreinte de déploiement réduite',
      s3p:
        'Pas besoin d’expédier un interpréteur complet et de grosses piles runtime pour le cœur du service—images et artefacts plus légers, y compris en périphérie et en environnements contraints.',
      s4t: 'Capacité utilisateurs concurrents plus élevée',
      s4p:
        'Les goroutines et une empreinte mémoire de base plus faible conviennent aux connexions longues et au multi-tenant—souvent plus de sessions simultanées et d’opérations d’admin sur le même matériel.',
    },
    bottom: {
      title: 'Essayez et contribuez',
      lead: 'Sous licence Apache 2.0. Stars, tickets et pull requests sont les bienvenus.',
      commercial: 'Besoin d’une licence commerciale ? Contactez-nous.',
      github: 'GitHub : LightningRAG/LightningRAG',
      articles: 'Parcourir les articles (mis à jour au fil du temps)',
    },
    footer: {
      copy: '© Communauté open source LightningRAG. Tous droits réservés.',
      docs: 'Documentation',
      githubRepo: 'Dépôt GitHub',
      articles: 'Articles',
      license: 'Licence commerciale',
      sitemap: 'Plan du site',
    },
    ui: { langAria: 'Langue' },
  },
  de: {
    meta: {
      title:
        'LightningRAG — Hochperformante Full-Stack-RAG- und Entwicklungsplattform',
      description:
        'LightningRAG ist eine Full-Stack-RAG- und Admin-Basis mit Go (Gin) und Vue 3. Gegenüber typischen Python-Stacks bietet sie höhere Leistung, kompiliertes Deployment, kleinere Artefakte und stärkere Nebenläufigkeit.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,Full Stack,Admin,Wissensbasis,Open Source',
      ogTitle:
        'LightningRAG — Hochperformante Full-Stack-RAG- und Entwicklungsplattform',
      ogDescription:
        'Unternehmens-RAG- und Admin-Plattform mit Go und Vue. Schnell, als Binaries auslieferbar, kompakt und mit skalierbarer Parallelität.',
      jsonLdDescription:
        'Full-Stack-RAG- und Admin-Basis mit Gin und Vue 3: JWT, dynamische Routen, Rechte und Codegenerierung.',
    },
    a11y: {
      skipMain: 'Zum Hauptinhalt springen',
      navMain: 'Hauptnavigation',
      brandHome: 'LightningRAG Startseite',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Funktionen',
      advantages: 'Vorteile',
      articles: 'Artikel',
      docs: 'Dokumentation',
      license: 'Kommerzielle Lizenz',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Full-Stack-RAG-Plattform',
      title:
        'Eine Hochleistungsbasis für Enterprise-RAG und Business-Backends',
      sub:
        'LightningRAG bündelt Authentifizierung, dynamisches Routing, Wissensbasen und Agenten—damit Sie sich auf Fachlogik und Modellqualität konzentrieren, nicht auf Infrastruktur von Grund auf.',
      ctaSource: 'Quellcode auf GitHub ansehen',
      ctaDocs: 'Dokumentation lesen',
    },
    features: {
      title: 'Produkthighlights',
      lead:
        'Erweitern Sie eine bewährte Full-Stack-Architektur für RAG—von Wissensbasen bis Agenten-Orchestrierung—mit einem Stack und einem Liefermodell.',
      c1t: 'Full Stack, einheitlich',
      c1p:
        'Gin im Backend, Vue 3 im Frontend, plus JWT, dynamische Menüs, Casbin und Codegenerierung—Admin-Oberflächen und RAG-Pipelines schneller ausliefern.',
      c2t: 'RAG-bereite Workflows',
      c2p:
        'Retrieval-Augmented-Pipelines, in denen Wissen, Dokumente und Agenten in Nutzer und Berechtigungen integrieren—ideal für private Deployments und Anpassung.',
      c3t: 'Erweiterbar und integrierbar',
      c3p:
        'Modulare Erweiterungspunkte für Vektorspeicher, Modell-Services und Enterprise-SSO. Webseite und Docs können parallel mit Artikeln und Release Notes wachsen.',
    },
    adv: {
      title: 'Kernvorteile gegenüber typischen Python-Projekten',
      lead:
        'Ein in Go kompilierter Backend ergänzt interpretierte Python-Stacks bei Leistung, Auslieferungsform und Betrieb—besonders wenn Latenz, Parallelität und Deployment-Fußabdruck zählen.',
      s1t: 'Geschwindigkeit und Durchsatz',
      s1p:
        'Native Binaries starten schnell mit vorhersehbarer GC. Bei stark parallelisiertem Retrieval und API-Aggregation erhalten Sie oft stabilere Latenz und höheren Durchsatz—horizontales Skalieren wird einfacher.',
      s2t: 'Kompilierte Auslieferung und stärkerer Code-Schutz',
      s2p:
        'Liefern Sie eine oder wenige Binaries ohne Klartext-Quellbaum auf dem Server—oft besser zum Schutz von Implementierungsdetails als typische Python-Quelltext-Deployments.',
      s3t: 'Kleinerer Deployment-Fußabdruck',
      s3p:
        'Kein vollständiger Interpreter und keine großen Runtime-Stacks für den Kern nötig—schlankere Images und Artefakte, auch am Edge und in ressourcenarmen Umgebungen.',
      s4t: 'Mehr gleichzeitige Nutzer',
      s4p:
        'Goroutines und geringerer Speicher-Baseline eignen sich für langlebige Verbindungen und Mandantenfähigkeit—oft mehr gleichzeitige Sessions und Admin-Vorgänge auf derselben Hardware.',
    },
    bottom: {
      title: 'Ausprobieren und mitwirken',
      lead: 'Lizenziert unter Apache 2.0. Stars, Issues und Pull Requests willkommen.',
      commercial: 'Kommerzielle Lizenz benötigt? Kontaktieren Sie uns.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: 'Artikel durchsuchen (fortlaufend aktualisiert)',
    },
    footer: {
      copy: '© LightningRAG Open-Source-Community. Alle Rechte vorbehalten.',
      docs: 'Dokumentation',
      githubRepo: 'GitHub-Repository',
      articles: 'Artikel',
      license: 'Kommerzielle Lizenz',
      sitemap: 'Sitemap',
    },
    ui: { langAria: 'Sprache' },
  },
  ja: {
    meta: {
      title:
        'LightningRAG — 高性能フルスタック RAG・開発基盤',
      description:
        'LightningRAG は Go（Gin）と Vue 3 のフルスタック RAG／管理基盤です。一般的な Python スタックと比べ、高い性能、コンパイル配布、小さな成果物、強い同時実行性を実現します。',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,フルスタック,管理,知識ベース,オープンソース',
      ogTitle: 'LightningRAG — 高性能フルスタック RAG・開発基盤',
      ogDescription:
        'Go と Vue で構築するエンタープライズ向け RAG／管理基盤。高速、バイナリ配布、コンパクト、スケーラブルな同時処理。',
      jsonLdDescription:
        'Gin と Vue 3 によるフルスタック RAG／管理基盤。JWT、動的ルート、権限、コード生成に対応。',
    },
    a11y: {
      skipMain: 'メインコンテンツへスキップ',
      navMain: 'メインナビゲーション',
      brandHome: 'LightningRAG ホーム',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: '特徴',
      advantages: '強み',
      articles: '記事',
      docs: 'ドキュメント',
      license: '商用ライセンス',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · フルスタック RAG 基盤',
      title: 'エンタープライズ RAG と業務バックエンド向けの高性能基盤',
      sub:
        'LightningRAG は認証、動的ルーティング、ナレッジベース、エージェントを統合し、ビジネスロジックとモデル品質に集中できるようにします。インフラの再構築に時間を取られません。',
      ctaSource: 'GitHub でソースを見る',
      ctaDocs: 'ドキュメントを読む',
    },
    features: {
      title: '製品ハイライト',
      lead:
        '実績のあるフルスタック構成を RAG に拡張—ナレッジベースからエージェント運用まで、単一スタックで提供します。',
      c1t: 'フルスタック統合',
      c1p:
        'バックエンドは Gin、フロントは Vue 3。JWT、動的メニュー、Casbin、コード生成で管理画面と RAG パイプラインを迅速に構築。',
      c2t: 'RAG 向けワークフロー',
      c2p:
        '検索拡張生成パイプラインで知識・ドキュメント・エージェントをユーザーと権限に接続。オンプレやカスタマイズに適しています。',
      c3t: '拡張と連携',
      c3p:
        'ベクトルストア、モデルサービス、企業 SSO などモジュール接続が容易。サイトとドキュメントを記事やリリースノートと並行して進化できます。',
    },
    adv: {
      title: '一般的な Python プロジェクトとの比較で得られる強み',
      lead:
        'Go によるコンパイル型バックエンドは、解釈型 Python スタックと性能・配布形態・運用面で補完関係にあり、遅延・同時実行・配布サイズが重要な場面で効きます。',
      s1t: '速度とスループット',
      s1p:
        'ネイティブバイナリは起動が速く GC も予測しやすい。高同時の検索や API 集約では遅延が安定しスループットが上がりやすく、水平スケールもしやすくなります。',
      s2t: 'コンパイル配布とコード保護',
      s2p:
        'ソースツリーを平文で載せず少数のバイナリで提供でき、実装詳細の保護には Python ソース配布より有利なことが多いです。',
      s3t: 'デプロイの小ささ',
      s3p:
        'コアサービスにフルインタプリタや巨大ランタイム一式が不要で、イメージと成果物を軽量化。エッジやリソース制限環境にも向きます。',
      s4t: 'より多くの同時ユーザー',
      s4p:
        'ゴルーチンと低いメモリ基盤は長寿命接続やマルチテナントに適し、同じハードウェアでより多くのセッションと管理操作を扱いやすくなります。',
    },
    bottom: {
      title: '試す・参加する',
      lead: 'Apache 2.0 ライセンス。Star、Issue、Pull Request を歓迎します。',
      commercial: '商用ライセンスが必要ですか？お問い合わせください。',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: '記事を見る（随時更新）',
    },
    footer: {
      copy: '© LightningRAG オープンソースコミュニティ。無断転載を禁じます。',
      docs: 'ドキュメント',
      githubRepo: 'GitHub リポジトリ',
      articles: '記事',
      license: '商用ライセンス',
      sitemap: 'サイトマップ',
    },
    ui: { langAria: '言語' },
  },
  ko: {
    meta: {
      title: 'LightningRAG — 고성능 풀스택 RAG 및 개발 기반 플랫폼',
      description:
        'LightningRAG는 Go(Gin)와 Vue 3 기반의 풀스택 RAG·관리 기반입니다. 일반적인 Python 스택 대비 높은 성능, 컴파일 배포, 작은 산출물, 강한 동시성을 제공합니다.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,풀스택,관리,지식베이스,오픈소스',
      ogTitle: 'LightningRAG — 고성능 풀스택 RAG 및 개발 기반 플랫폼',
      ogDescription:
        'Go와 Vue로 구축하는 엔터프라이즈 RAG·관리 플랫폼. 빠르고, 바이너리로 배포 가능하며, 컴팩트하고 동시 처리 확장에 유리합니다.',
      jsonLdDescription:
        'Gin과 Vue 3 기반 풀스택 RAG·관리 기반. JWT, 동적 라우트, 권한, 코드 생성 지원.',
    },
    a11y: {
      skipMain: '본문으로 건너뛰기',
      navMain: '주요 탐색',
      brandHome: 'LightningRAG 홈',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: '특징',
      advantages: '장점',
      articles: '글',
      docs: '문서',
      license: '상용 라이선스',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · 풀스택 RAG 플랫폼',
      title: '엔터프라이즈 RAG와 비즈니스 백엔드를 위한 고성능 기반',
      sub:
        'LightningRAG는 인증, 동적 라우팅, 지식베이스, 에이전트를 한데 모아 비즈니스 로직과 모델 품질에 집중할 수 있게 합니다. 인프라를 처음부터 다시 짓지 않아도 됩니다.',
      ctaSource: 'GitHub에서 소스 보기',
      ctaDocs: '문서 읽기',
    },
    features: {
      title: '제품 하이라이트',
      lead:
        '검증된 풀스택 아키텍처를 RAG로 확장—지식베이스부터 에이전트 운영까지 단일 스택으로 제공합니다.',
      c1t: '풀스택 일체화',
      c1p:
        '백엔드 Gin, 프론트 Vue 3, JWT·동적 메뉴·Casbin·코드 생성으로 관리 콘솔과 RAG 파이프라인을 빠르게 구축합니다.',
      c2t: 'RAG 준비 워크플로',
      c2p:
        '검색 증강 생성 파이프라인에서 지식·문서·에이전트를 사용자·권한과 연결. 폐쇄 배포와 맞춤화에 적합합니다.',
      c3t: '확장·연동',
      c3p:
        '벡터 스토어, 모델 서비스, 기업 SSO 등 모듈형 연동. 사이트와 문서를 글·릴리스 노트와 함께 성장시킬 수 있습니다.',
    },
    adv: {
      title: '일반적인 Python 프로젝트 대비 핵심 이점',
      lead:
        'Go 컴파일 백엔드는 해석형 Python 스택과 성능·배포 형태·운영 측면에서 상호 보완적이며, 지연·동시성·배포 크기가 중요할 때 특히 빛을 발합니다.',
      s1t: '속도와 처리량',
      s1p:
        '네이티브 바이너리는 빠르게 기동하고 GC 예측이 쉽습니다. 고동시 검색·API 집계에서 지연이 안정되고 처리량이 높아지며 수평 확장이 수월합니다.',
      s2t: '컴파일 배포와 코드 보호',
      s2p:
        '소스 트리를 평문으로 노출하지 않고 소수의 바이너리로 제공해 구현 세부를 보호하기에 Python 소스 배포보다 유리한 경우가 많습니다.',
      s3t: '배포 부담 축소',
      s3p:
        '핵심 서비스에 전체 인터프리터와 거대 런타임 스택이 필요 없어 이미지와 산출물이 가벼워지고 엣지·자원 제한 환경에도 적합합니다.',
      s4t: '더 많은 동시 사용자',
      s4p:
        '고루틴과 낮은 메모리 기반은 장기 연결·멀티 테넌트에 적합해 동일 하드웨어에서 더 많은 세션과 관리 작업을 처리하기 쉽습니다.',
    },
    bottom: {
      title: '사용해 보고 기여하기',
      lead: 'Apache 2.0 라이선스. Star, Issue, Pull Request를 환영합니다.',
      commercial: '상용 라이선스가 필요하신가요? 문의해 주세요.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: '글 보기(지속 업데이트)',
    },
    footer: {
      copy: '© LightningRAG 오픈소스 커뮤니티. 무단 복제를 금합니다.',
      docs: '문서',
      githubRepo: 'GitHub 저장소',
      articles: '글',
      license: '상용 라이선스',
      sitemap: '사이트맵',
    },
    ui: { langAria: '언어' },
  },
  pt: {
    meta: {
      title:
        'LightningRAG — Plataforma RAG e de desenvolvimento full stack de alto desempenho',
      description:
        'LightningRAG é uma base RAG e administrativa full stack em Go (Gin) e Vue 3. Em relação a pilhas Python típicas, oferece maior desempenho, implantação compilada, artefatos menores e concorrência mais forte.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,full stack,admin,base de conhecimento,código aberto',
      ogTitle:
        'LightningRAG — Plataforma RAG e de desenvolvimento full stack de alto desempenho',
      ogDescription:
        'Plataforma RAG e administrativa empresarial com Go e Vue. Rápida, entregável como binários, compacta e com concorrência escalável.',
      jsonLdDescription:
        'Base RAG e administrativa full stack com Gin e Vue 3: JWT, rotas dinâmicas, permissões e geração de código.',
    },
    a11y: {
      skipMain: 'Ir para o conteúdo principal',
      navMain: 'Navegação principal',
      brandHome: 'Página inicial LightningRAG',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Recursos',
      advantages: 'Vantagens',
      articles: 'Artigos',
      docs: 'Documentação',
      license: 'Licença comercial',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Plataforma RAG full stack',
      title:
        'Uma base de alto desempenho para RAG corporativo e backends de negócio',
      sub:
        'LightningRAG reúne autenticação, roteamento dinâmico, bases de conhecimento e agentes para você focar na lógica de negócio e na qualidade do modelo—não em reconstruir infraestrutura.',
      ctaSource: 'Ver código no GitHub',
      ctaDocs: 'Ler a documentação',
    },
    features: {
      title: 'Destaques do produto',
      lead:
        'Estenda uma arquitetura full stack comprovada para RAG—de bases de conhecimento à orquestração de agentes—com uma única pilha e modelo de entrega.',
      c1t: 'Full stack unificado',
      c1p:
        'Gin no backend e Vue 3 no frontend, além de JWT, menus dinâmicos, Casbin e geração de código—entregue consoles admin e pipelines RAG mais rápido.',
      c2t: 'Fluxos prontos para RAG',
      c2p:
        'Pipelines de recuperação aumentada em que conhecimento, documentos e agentes se integram a usuários e permissões—ideal para implantações privadas e personalização.',
      c3t: 'Extensível e integrável',
      c3p:
        'Ganchos modulares para armazenamentos vetoriais, serviços de modelos e SSO corporativo. Site e documentação evoluem junto com artigos e notas de versão.',
    },
    adv: {
      title: 'Vantagens em relação a projetos Python típicos',
      lead:
        'Um backend compilado em Go complementa pilhas Python interpretadas em desempenho, modelo de entrega e operações—especialmente quando latência, concorrência e tamanho da implantação importam.',
      s1t: 'Velocidade e vazão',
      s1p:
        'Binários nativos iniciam rápido com GC previsível. Em recuperação altamente concorrente e agregação de APIs, você costuma obter latência mais estável e maior vazão—com escalonamento horizontal mais simples.',
      s2t: 'Entrega compilada e melhor proteção de código',
      s2p:
        'Entregue um ou poucos binários sem expor a árvore de fontes em texto puro no servidor—muitas vezes melhor para proteger detalhes de implementação do que implantações Python a partir de código-fonte.',
      s3t: 'Implantação mais enxuta',
      s3p:
        'Sem necessidade de enviar um interpretador completo e grandes pilhas de runtime para o serviço principal—imagens e artefatos mais leves, inclusive na borda e em ambientes com recursos limitados.',
      s4t: 'Maior capacidade de usuários simultâneos',
      s4p:
        'Goroutines e menor uso base de memória combinam com conexões longas e cenários multi-inquilino—muitas vezes mais sessões simultâneas e operações administrativas no mesmo hardware.',
    },
    bottom: {
      title: 'Experimente e contribua',
      lead: 'Licenciado sob Apache 2.0. Stars, issues e pull requests são bem-vindos.',
      commercial: 'Precisa de licença comercial? Fale conosco.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: 'Ver artigos (atualizados ao longo do tempo)',
    },
    footer: {
      copy: '© Comunidade open source LightningRAG. Todos os direitos reservados.',
      docs: 'Documentação',
      githubRepo: 'Repositório no GitHub',
      articles: 'Artigos',
      license: 'Licença comercial',
      sitemap: 'Mapa do site',
    },
    ui: { langAria: 'Idioma' },
  },
  ru: {
    meta: {
      title:
        'LightningRAG — высокопроизводительная full-stack платформа RAG и разработки',
      description:
        'LightningRAG — full-stack основа для RAG и админки на Go (Gin) и Vue 3. По сравнению с типичными Python-стеками даёт более высокую производительность, компиляционный деплой, меньшие артефакты и сильнее нагрузку по параллелизму.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,full stack,админка,база знаний,открытый код',
      ogTitle:
        'LightningRAG — высокопроизводительная full-stack платформа RAG и разработки',
      ogDescription:
        'Корпоративная платформа RAG и админки на Go и Vue. Быстро, поставляется бинарниками, компактно и с масштабируемой параллельностью.',
      jsonLdDescription:
        'Full-stack основа RAG и админки на Gin и Vue 3: JWT, динамические маршруты, права и генерация кода.',
    },
    a11y: {
      skipMain: 'Перейти к основному содержимому',
      navMain: 'Основная навигация',
      brandHome: 'Главная LightningRAG',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Возможности',
      advantages: 'Преимущества',
      articles: 'Статьи',
      docs: 'Документация',
      license: 'Коммерческая лицензия',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Full-stack платформа RAG',
      title:
        'Высокопроизводительная основа для корпоративного RAG и бизнес-бэкендов',
      sub:
        'LightningRAG объединяет аутентификацию, динамический роутинг, базы знаний и агентов — чтобы вы сосредоточились на бизнес-логике и качестве моделей, а не на построении инфраструктуры с нуля.',
      ctaSource: 'Исходный код на GitHub',
      ctaDocs: 'Читать документацию',
    },
    features: {
      title: 'Ключевые особенности',
      lead:
        'Расширяйте проверенную full-stack архитектуру для RAG — от баз знаний до оркестрации агентов — единым стеком и моделью поставки.',
      c1t: 'Единый full-stack',
      c1p:
        'Gin на бэкенде и Vue 3 на фронтенде, плюс JWT, динамические меню, Casbin и генерация кода — быстрее доставляйте админки и RAG-пайплайны.',
      c2t: 'Готовые сценарии RAG',
      c2p:
        'Пайплайны retrieval-augmented generation, где знания, документы и агенты интегрируются с пользователями и правами — удобно для частных развёртываний и кастомизации.',
      c3t: 'Расширяемость и интеграции',
      c3p:
        'Модульные точки расширения для векторных хранилищ, сервисов моделей и корпоративного SSO. Сайт и документация развиваются вместе со статьями и релиз-нотами.',
    },
    adv: {
      title: 'Ключевые преимущества перед типичными Python-проектами',
      lead:
        'Скомпилированный бэкенд на Go дополняет интерпретируемые Python-стеки по производительности, модели поставки и эксплуатации — особенно когда важны задержка, параллелизм и размер деплоя.',
      s1t: 'Скорость и пропускная способность',
      s1p:
        'Нативные бинарники быстро стартуют с предсказуемым GC. При высокой параллельной выдаче и агрегации API обычно стабильнее задержка и выше throughput — проще масштабировать горизонтально.',
      s2t: 'Компиляционная поставка и защита кода',
      s2p:
        'Поставляйте один или несколько бинарников без открытого дерева исходников на сервере — часто лучше для защиты деталей реализации, чем типичный деплой Python из исходников.',
      s3t: 'Меньший объём деплоя',
      s3p:
        'Не нужен полный интерпретатор и тяжёлые рантайм-стеки для ядра сервиса — легче образы и артефакты, в том числе на периферии и в условиях ограниченных ресурсов.',
      s4t: 'Больше одновременных пользователей',
      s4p:
        'Горутины и меньший базовый расход памяти подходят для долгих соединений и мультитенантности — часто больше одновременных сессий и админ-операций на том же железе.',
    },
    bottom: {
      title: 'Попробуйте и участвуйте',
      lead: 'Лицензия Apache 2.0. Приветствуются звёзды, issues и pull request\'ы.',
      commercial: 'Нужна коммерческая лицензия? Свяжитесь с нами.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: 'Статьи (обновляются со временем)',
    },
    footer: {
      copy: '© Сообщество LightningRAG с открытым исходным кодом. Все права защищены.',
      docs: 'Документация',
      githubRepo: 'Репозиторий на GitHub',
      articles: 'Статьи',
      license: 'Коммерческая лицензия',
      sitemap: 'Карта сайта',
    },
    ui: { langAria: 'Язык' },
  },
  it: {
    meta: {
      title:
        'LightningRAG — Piattaforma RAG e di sviluppo full stack ad alte prestazioni',
      description:
        'LightningRAG è una base RAG e amministrativa full stack con Go (Gin) e Vue 3. Rispetto agli stack Python tipici offre prestazioni superiori, deployment compilato, artefatti più piccoli e maggiore concorrenza.',
      keywords:
        'LightningRAG,RAG,Gin,Vue 3,Go,full stack,admin,base di conoscenza,open source',
      ogTitle:
        'LightningRAG — Piattaforma RAG e di sviluppo full stack ad alte prestazioni',
      ogDescription:
        'Piattaforma RAG e amministrativa enterprise con Go e Vue. Veloce, distribuibile come binari, compatta e con concorrenza scalabile.',
      jsonLdDescription:
        'Base RAG e amministrativa full stack con Gin e Vue 3: JWT, route dinamiche, permessi e generazione di codice.',
    },
    a11y: {
      skipMain: 'Vai al contenuto principale',
      navMain: 'Navigazione principale',
      brandHome: 'Home LightningRAG',
    },
    brand: { logoAlt: 'LightningRAG' },
    nav: {
      features: 'Funzionalità',
      advantages: 'Vantaggi',
      articles: 'Articoli',
      docs: 'Documentazione',
      license: 'Licenza commerciale',
    },
    btn: { github: 'GitHub' },
    hero: {
      badge: 'Go · Gin · Vue 3 · Piattaforma RAG full stack',
      title:
        'Una base ad alte prestazioni per RAG enterprise e backend di business',
      sub:
        'LightningRAG unisce autenticazione, routing dinamico, basi di conoscenza e agenti così puoi concentrarti sulla logica di business e sulla qualità del modello—non sulla ricostruzione dell’infrastruttura.',
      ctaSource: 'Vedi il codice su GitHub',
      ctaDocs: 'Leggi la documentazione',
    },
    features: {
      title: 'Punti salienti',
      lead:
        'Estendi un’architettura full stack collaudata per il RAG—dalle basi di conoscenza all’orchestrazione degli agenti—con uno stack e un modello di consegna unificati.',
      c1t: 'Full stack unificato',
      c1p:
        'Gin nel backend e Vue 3 nel frontend, più JWT, menu dinamici, Casbin e generazione di codice—consegna più velocemente console admin e pipeline RAG.',
      c2t: 'Flussi pronti per il RAG',
      c2p:
        'Pipeline retrieval-augmented in cui conoscenza, documenti e agenti si integrano con utenti e permessi—ideale per deployment privati e personalizzazione.',
      c3t: 'Estensibile e integrabile',
      c3p:
        'Hook modulari per archivi vettoriali, servizi di modelli e SSO enterprise. Sito e documentazione evolvono insieme ad articoli e note di rilascio.',
    },
    adv: {
      title: 'Vantaggi rispetto ai progetti Python tipici',
      lead:
        'Un backend compilato in Go completa gli stack Python interpretati su prestazioni, modello di consegna e operazioni—soprattutto quando contano latenza, concorrenza e impronta di deployment.',
      s1t: 'Velocità e throughput',
      s1p:
        'I binari nativi partono velocemente con GC prevedibile. Con recupero ad alta concorrenza e aggregazione API si ottiene spesso latenza più stabile e throughput maggiore—con scale-out orizzontale più semplice.',
      s2t: 'Consegna compilata e migliore protezione del codice',
      s2p:
        'Consegna uno o pochi binari senza esporre l’albero dei sorgenti in chiaro sul server—spesso meglio per proteggere i dettagli di implementazione rispetto ai deployment Python da sorgente.',
      s3t: 'Impronta di deployment più piccola',
      s3p:
        'Non serve spedire un interprete completo e grandi stack runtime per il servizio core—immagini e artefatti più leggeri, anche al edge e in ambienti con risorse limitate.',
      s4t: 'Maggiore capacità di utenti concorrenti',
      s4p:
        'Le goroutine e una baseline di memoria inferiore si adattano a connessioni long-lived e scenari multi-tenant—spesso più sessioni simultanee e operazioni admin sulla stessa hardware.',
    },
    bottom: {
      title: 'Prova e contribuisci',
      lead: 'Licenza Apache 2.0. Benvenuti star, issue e pull request.',
      commercial: 'Serve una licenza commerciale? Contattaci.',
      github: 'GitHub: LightningRAG/LightningRAG',
      articles: 'Sfoglia gli articoli (aggiornati nel tempo)',
    },
    footer: {
      copy: '© Comunità open source LightningRAG. Tutti i diritti riservati.',
      docs: 'Documentazione',
      githubRepo: 'Repository GitHub',
      articles: 'Articoli',
      license: 'Licenza commerciale',
      sitemap: 'Mappa del sito',
    },
    ui: { langAria: 'Lingua' },
  },
};
