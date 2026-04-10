/**
 * 文档站文案（与 docs-i18n.js 配合）。结构：共用 nav/footer + 各页 meta 与正文键。
 * 非英语语种在保持语义一致的前提下适当精简句式。
 */

function deepMerge(base, patch) {
  const out = { ...base };
  for (const k of Object.keys(patch)) {
    const pv = patch[k];
    const bv = base[k];
    if (
      pv != null &&
      typeof pv === 'object' &&
      !Array.isArray(pv) &&
      bv != null &&
      typeof bv === 'object' &&
      !Array.isArray(bv)
    ) {
      out[k] = deepMerge(bv, pv);
    } else {
      out[k] = pv;
    }
  }
  return out;
}

const EN = {
  metaHub: {
    title: 'Documentation — LightningRAG',
    description:
      'LightningRAG documentation: quick start, local development, RAG knowledge bases, models, chat, and agents.',
  },
  metaInit: {
    title: 'Quick start — LightningRAG',
    description:
      'Set up Node.js and Go, run the Gin API and Vue 3 admin, and open Swagger for LightningRAG.',
  },
  metaUse: {
    title: 'Using RAG features — LightningRAG',
    description:
      'How to use knowledge bases, model providers, RAG chat with citations, and agent workflows in LightningRAG.',
  },
  metaLicense: {
    title: 'Commercial license — LightningRAG',
    description:
      'Purchase a commercial license or request a quote: contact LightningRAG for enterprise terms and branding notices.',
  },
  metaPreview: {
    title: 'Product interface — LightningRAG',
    description:
      'Screenshots of the LightningRAG admin console: knowledge bases, retrieval, chat, models, RAG settings, and agent orchestration.',
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
  breadcrumb: {
    hub: 'Documentation',
    init: 'Quick start',
    using: 'RAG & agents',
    preview: 'Product UI',
  },
  btn: { github: 'GitHub' },
  ui: { langAria: 'Language' },
  footer: {
    copy: '© LightningRAG open-source community. All rights reserved.',
    home: 'Home',
    docs: 'Documentation',
    articles: 'Articles',
    license: 'Commercial license',
    sitemap: 'Sitemap',
    githubRepo: 'GitHub repository',
  },
  docHub: {
    title: 'Documentation',
    lead:
      'Learn how to run LightningRAG locally, then use knowledge bases, models, conversational RAG, and agent flows in the admin console.',
    c1t: 'Quick start',
    c1p:
      'Prerequisites, clone the repo, run the Go server and Vue frontend, and optional Swagger / Docker paths.',
    c1a: 'Open quick start →',
    c2t: 'RAG features & agents',
    c2p:
      'Knowledge bases, embedding and chat models, retrieval Q&A with citations, and visual agent orchestration.',
    c2a: 'Open usage guide →',
    c3t: 'Source & license',
    c3p:
      'Apache 2.0. See the README for commercial terms, contributing, and the full feature list.',
    c3a: 'View on GitHub →',
    c3license: 'Commercial license & contact →',
    c4t: 'Product interface',
    c4p:
      'Screenshots of the admin console: knowledge bases, document lists, retrieval, AI chat, models, RAG config, and agent flows.',
    c4a: 'View product UI →',
  },
  pageLicense: {
    title: 'Commercial license',
    lead:
      'LightningRAG is open source under Apache License 2.0. Some production uses may require a commercial license—for example to comply with branding or attribution requirements described in the project notices.',
    p1:
      'To purchase a license, request a quote, or discuss enterprise terms, contact us using the email below. Please include your company or organization name, approximate deployment scale, and how you plan to use LightningRAG.',
    hContact: 'Contact',
    emailIntro: 'Sales & licensing:',
    emailNote: 'We aim to respond within a few business days.',
    backHome: '← Back to home',
  },
  pageInit: {
    title: 'Quick start',
    lead:
      'LightningRAG is a Go (Gin) + Vue 3 full-stack project. Run the API and admin UI on your machine, then explore APIs via Swagger.',
    prereqH: 'Prerequisites',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git, and an IDE (GoLand or VS Code recommended)',
    serverH: 'Backend (server)',
    s1:
      'Open the server directory as the project root in your IDE (not the repository root). Then install tools and run:',
    sNote:
      'Default API listens on port 8888. Configure databases and Redis via config files under server/ as needed.',
    webH: 'Frontend (web)',
    w1:
      'From the web directory, install dependencies and start the dev server. Point the UI at your local API base URL in environment or proxy settings if required.',
    swaggerH: 'Swagger API docs',
    sw1:
      'Install swag, run swag init inside server/, restart the API, then open /swagger/index.html on the API host.',
    dockerH: 'Docker & compose',
    d1:
      'For containerized dependencies or full stacks, see deploy/docker-compose in the repository. Adjust profiles and .env for your vector store and services.',
    moreH: 'Next steps',
    m1:
      'Read the repository README for Casbin menus, code generation, plugin layout, and RAG module details. Commercial use follows Apache 2.0 with the notices described there.',
    m2license: 'Commercial license & contact →',
  },
  pageUse: {
    title: 'Using RAG features',
    lead:
      'The admin console integrates knowledge bases, model configuration, chat with retrieval, and agent designer. Below is a concise map of where to click and what to configure.',
    kbH: 'Knowledge bases',
    kb1:
      'Create a knowledge base, upload or import documents, choose chunking and parsing options, and wait for indexing to finish.',
    kb2:
      'Bind embedding and optional rerank models per KB or use system defaults so retrieval uses your configured vector store.',
    modelH: 'Models & providers',
    mo1:
      'Configure LLM, embedding, rerank, and other providers under system settings. Many OpenAI-compatible endpoints and local runtimes (e.g. Ollama) are supported via the registry pattern.',
    chatH: 'Conversational RAG',
    ch1:
      'Start a RAG conversation from the chat UI: select knowledge bases, ask questions, and review inline citations that map to retrieved chunks.',
    agentH: 'Agents & workflows',
    ag1:
      'Use the agent canvas to compose flows from templates or scratch: Begin, Retrieval, LLM, Message, branches, HTTP tools, and more. Save versions and run with user queries.',
    securityH: 'Users & permissions',
    sec1:
      'JWT authentication, dynamic routes, and Casbin-backed authorization align menu visibility and APIs with roles—reuse this for multi-tenant or internal deployments.',
  },
  pagePreview: {
    title: 'Product interface',
    lead:
      'Screenshots from the Vue 3 admin console. Exact layout may vary slightly by release.',
    cap1: 'Knowledge bases',
    cap2: 'Knowledge base document list',
    cap3: 'Document retrieval',
    cap4: 'AI chat',
    cap5: 'Model configuration',
    cap6: 'RAG system settings',
    cap7: 'Agent orchestration canvas',
    cap8: 'Agent list in orchestration',
    alt1: 'Screenshot: knowledge bases in the admin console',
    alt2: 'Screenshot: document list inside a knowledge base',
    alt3: 'Screenshot: document retrieval interface',
    alt4: 'Screenshot: AI conversation interface',
    alt5: 'Screenshot: model provider configuration',
    alt6: 'Screenshot: RAG system configuration',
    alt7: 'Screenshot: visual agent workflow canvas',
    alt8: 'Screenshot: agent list in the orchestration module',
  },
};

const ZH_CN = {
  metaHub: {
    title: '文档 — LightningRAG',
    description:
      'LightningRAG 使用文档：快速开始、本地开发、知识库、模型、对话与智能体编排。',
  },
  metaInit: {
    title: '快速开始 — LightningRAG',
    description:
      '配置 Node.js 与 Go，启动 Gin 后端与 Vue 管理端，并通过 Swagger 浏览 API。',
  },
  metaUse: {
    title: '使用 RAG 能力 — LightningRAG',
    description:
      '知识库管理、模型接入、带引用的检索对话，以及智能体工作流的使用说明。',
  },
  metaLicense: {
    title: '商用授权 — LightningRAG',
    description:
      '购买商用授权或索取报价：通过下列方式联系 LightningRAG，了解企业条款与品牌标识相关约定。',
  },
  metaPreview: {
    title: '产品界面 — LightningRAG',
    description:
      'LightningRAG 管理后台界面截图：知识库、检索、对话、模型、RAG 配置与智能体编排。',
  },
  a11y: {
    skipMain: '跳到主要内容',
    navMain: '主导航',
    brandHome: 'LightningRAG 首页',
  },
  nav: {
    features: '产品特点',
    advantages: '核心优势',
    articles: '文章',
    docs: '文档',
    license: '商用授权',
  },
  breadcrumb: {
    hub: '文档',
    init: '快速开始',
    using: 'RAG 与智能体',
    preview: '产品界面',
  },
  ui: { langAria: '语言' },
  footer: {
    copy: '© LightningRAG 开源社区。保留所有权利。',
    home: '首页',
    docs: '文档',
    articles: '文章',
    license: '商用授权',
    sitemap: '站点地图',
    githubRepo: 'GitHub 仓库',
  },
  docHub: {
    title: '文档中心',
    lead:
      '了解如何在本地运行 LightningRAG，并在管理后台中使用知识库、模型、检索对话与智能体编排。',
    c1t: '快速开始',
    c1p: '环境要求、克隆仓库、启动 Go 服务与 Vue 前端，以及 Swagger / Docker 相关说明。',
    c1a: '进入快速开始 →',
    c2t: 'RAG 与智能体',
    c2p: '知识库、嵌入与对话模型、带引用问答，以及可视化智能体流程。',
    c2a: '进入使用说明 →',
    c3t: '源码与协议',
    c3p: 'Apache 2.0。商业使用与贡献方式见仓库 README。',
    c3a: '在 GitHub 上查看 →',
    c3license: '商用授权与联系 →',
    c4t: '产品界面',
    c4p: '管理后台截图：知识库、文档列表、检索、AI 对话、模型、RAG 配置与智能体流程。',
    c4a: '查看产品界面 →',
  },
  pageLicense: {
    title: '商用授权',
    lead:
      'LightningRAG 以 Apache License 2.0 开源。部分生产环境使用可能需要另行取得商用授权，例如为满足项目中关于品牌展示或署名等声明的要求。',
    p1:
      '如需购买授权、索取报价或洽谈企业条款，请通过下方邮箱联系。建议在邮件中说明公司或组织名称、大致部署规模以及使用场景。',
    hContact: '联系方式',
    emailIntro: '销售与授权咨询：',
    wechatIntro: '微信客服：',
    wechatLink: '联系企业微信客服',
    emailNote: '我们一般在数个工作日内回复。',
    backHome: '← 返回首页',
  },
  pageInit: {
    title: '快速开始',
    lead:
      'LightningRAG 为 Go（Gin）+ Vue 3 全栈项目。在本地启动 API 与管理界面后，可通过 Swagger 浏览接口。',
    prereqH: '环境要求',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git；推荐使用 GoLand 或 VS Code',
    serverH: '后端（server）',
    s1:
      '在 IDE 中以 server 目录作为工程根目录打开（不要直接以仓库根目录作为 Go 工程根）。安装工具并运行：',
    sNote:
      '默认 API 监听 8888 端口。数据库、Redis 等请在 server 下配置文件中按需修改。',
    webH: '前端（web）',
    w1:
      '在 web 目录安装依赖并启动开发服务器。若需自定义接口地址，请在环境变量或代理配置中指向本地 API。',
    swaggerH: 'Swagger 接口文档',
    sw1:
      '安装 swag，在 server 目录执行 swag init，重启服务后访问 API 地址下的 /swagger/index.html。',
    dockerH: 'Docker 与 Compose',
    d1:
      '容器化依赖或整套环境见仓库 deploy/docker-compose，按向量库等需求调整 profile 与 .env。',
    moreH: '后续阅读',
    m1:
      '完整能力（动态菜单、代码生成、插件、RAG 模块细节）见仓库 README；商用请遵守 Apache 2.0 及其中声明。',
    m2license: '商用授权与联系 →',
  },
  pageUse: {
    title: '使用 RAG 相关能力',
    lead:
      '管理后台整合了知识库、模型配置、检索对话与智能体设计器。以下为功能入口与配置要点概览。',
    kbH: '知识库',
    kb1: '新建知识库、上传或导入文档，选择解析与分块策略，等待索引完成。',
    kb2: '为知识库绑定嵌入与可选重排序模型，或使用系统默认，以便写入配置的向量存储。',
    modelH: '模型与供应商',
    mo1:
      '在系统设置中配置 LLM、嵌入、重排等；多数 OpenAI 兼容接口与本地推理（如 Ollama）可通过注册机制接入。',
    chatH: '检索增强对话',
    ch1: '在对话界面选择知识库提问，查看与检索片段对应的正文内引用角标。',
    agentH: '智能体与工作流',
    ag1:
      '在画布中从模板或空白搭建流程：Begin、Retrieval、LLM、Message、分支与 HTTP 等组件，保存版本并以用户问题运行。',
    securityH: '用户与权限',
    sec1:
      'JWT、动态路由与 Casbin 权限控制菜单与接口可见性，适合私有化与多角色场景。',
  },
  pagePreview: {
    title: '产品界面',
    lead: '以下为 Vue 3 管理后台界面截图，具体布局可能随版本略有差异。',
    cap1: '知识库',
    cap2: '知识库文档列表',
    cap3: '文档检索',
    cap4: 'AI 对话',
    cap5: '模型配置管理',
    cap6: 'RAG 系统配置',
    cap7: '智能体编排画布',
    cap8: '智能体编排中的 Agent 列表',
    alt1: '截图：知识库管理界面',
    alt2: '截图：知识库内文档列表',
    alt3: '截图：文档检索界面',
    alt4: '截图：AI 对话界面',
    alt5: '截图：模型配置管理',
    alt6: '截图：RAG 系统配置',
    alt7: '截图：智能体编排画布',
    alt8: '截图：智能体编排 Agent 列表',
  },
};

const ZH_TW = {
  metaHub: {
    title: '文件 — LightningRAG',
    description:
      'LightningRAG 說明文件：快速開始、本機開發、知識庫、模型、對話與智慧體編排。',
  },
  metaInit: {
    title: '快速開始 — LightningRAG',
    description:
      '設定 Node.js 與 Go，啟動 Gin 後端與 Vue 管理介面，並以 Swagger 瀏覽 API。',
  },
  metaUse: {
    title: '使用 RAG 功能 — LightningRAG',
    description:
      '知識庫管理、模型串接、含引用的檢索對話，以及智慧體工作流程說明。',
  },
  metaLicense: {
    title: '商用授權 — LightningRAG',
    description:
      '購買商用授權或索取報價：透過下列方式聯絡 LightningRAG，了解企業條款與品牌標示相關約定。',
  },
  metaPreview: {
    title: '產品介面 — LightningRAG',
    description:
      'LightningRAG 管理後台介面截圖：知識庫、檢索、對話、模型、RAG 設定與智慧體編排。',
  },
  a11y: {
    skipMain: '跳到主要內容',
    navMain: '主要導覽',
    brandHome: 'LightningRAG 首頁',
  },
  nav: {
    features: '產品特點',
    advantages: '核心優勢',
    articles: '文章',
    docs: '文件',
    license: '商用授權',
  },
  breadcrumb: {
    hub: '文件',
    init: '快速開始',
    using: 'RAG 與智慧體',
    preview: '產品介面',
  },
  ui: { langAria: '語言' },
  footer: {
    copy: '© LightningRAG 開源社群。保留所有權利。',
    home: '首頁',
    docs: '文件',
    articles: '文章',
    license: '商用授權',
    sitemap: '網站地圖',
    githubRepo: 'GitHub 儲存庫',
  },
  docHub: {
    title: '文件中心',
    lead:
      '了解如何在本機執行 LightningRAG，並於管理後台使用知識庫、模型、檢索對話與智慧體編排。',
    c1t: '快速開始',
    c1p: '環境需求、複製儲存庫、啟動 Go 服務與 Vue 前端，以及 Swagger／Docker 說明。',
    c1a: '前往快速開始 →',
    c2t: 'RAG 與智慧體',
    c2p: '知識庫、嵌入與對話模型、含引用問答，以及視覺化智慧體流程。',
    c2a: '前往使用說明 →',
    c3t: '原始碼與授權',
    c3p: 'Apache 2.0。商業使用與貢獻方式見儲存庫 README。',
    c3a: '在 GitHub 上檢視 →',
    c3license: '商用授權與聯絡 →',
    c4t: '產品介面',
    c4p: '管理後台截圖：知識庫、文件列表、檢索、AI 對話、模型、RAG 設定與智慧體流程。',
    c4a: '檢視產品介面 →',
  },
  pageLicense: {
    title: '商用授權',
    lead:
      'LightningRAG 以 Apache License 2.0 開源。部分正式環境使用可能需要另行取得商用授權，例如為符合專案中關於品牌展示或署名等聲明的要求。',
    p1:
      '如需購買授權、索取報價或洽談企業條款，請透過下方電子郵件聯絡。建議於郵件中說明公司或組織名稱、大致部署規模以及使用情境。',
    hContact: '聯絡方式',
    emailIntro: '銷售與授權諮詢：',
    emailNote: '我們一般在數個工作天內回覆。',
    backHome: '← 返回首頁',
  },
  pageInit: {
    title: '快速開始',
    lead:
      'LightningRAG 為 Go（Gin）+ Vue 3 全端專案。於本機啟動 API 與管理介面後，可透過 Swagger 瀏覽介面。',
    prereqH: '環境需求',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git；建議使用 GoLand 或 VS Code',
    serverH: '後端（server）',
    s1:
      '請在 IDE 以 server 目錄作為專案根目錄（勿直接以儲存庫根目錄作為 Go 專案根）。安裝工具並執行：',
    sNote:
      '預設 API 監聽 8888。資料庫、Redis 等請於 server 下設定檔依需求調整。',
    webH: '前端（web）',
    w1:
      '於 web 目錄安裝相依並啟動開發伺服器。若需自訂 API 位址，請於環境變數或代理設定指向本機 API。',
    swaggerH: 'Swagger API 文件',
    sw1:
      '安裝 swag，於 server 目錄執行 swag init，重啟服務後開啟 /swagger/index.html。',
    dockerH: 'Docker 與 Compose',
    d1:
      '容器化相依或整套環境見儲存庫 deploy/docker-compose，依向量資料庫等需求調整 profile 與 .env。',
    moreH: '後續閱讀',
    m1:
      '完整能力見儲存庫 README；商業使用請遵守 Apache 2.0 與其中聲明。',
    m2license: '商用授權與聯絡 →',
  },
  pageUse: {
    title: '使用 RAG 相關功能',
    lead:
      '管理後台整合知識庫、模型設定、檢索對話與智慧體設計器。以下為功能入口與設定重點摘要。',
    kbH: '知識庫',
    kb1: '建立知識庫、上傳或匯入文件，選擇解析與分塊策略，等待索引完成。',
    kb2: '為知識庫綁定嵌入與可選重排序模型，或使用系統預設，以寫入所設定的向量儲存。',
    modelH: '模型與供應商',
    mo1:
      '於系統設定中設定 LLM、嵌入、重排等；多數 OpenAI 相容端點與本機推理可透過註冊機制接入。',
    chatH: '檢索增強對話',
    ch1: '於對話介面選擇知識庫提問，檢視對應检索片段的文中引用標記。',
    agentH: '智慧體與工作流程',
    ag1:
      '於畫布從範本或空白建立流程：Begin、Retrieval、LLM、Message、分支與 HTTP 等元件，儲存版本並以使用者問題執行。',
    securityH: '使用者與權限',
    sec1:
      'JWT、動態路由與 Casbin 權限對齊選單與 API 可見性，適合私有化與多角色情境。',
  },
  pagePreview: {
    title: '產品介面',
    lead: '以下為 Vue 3 管理後台介面截圖，版面可能隨版本略有不同。',
    cap1: '知識庫',
    cap2: '知識庫文件列表',
    cap3: '文件檢索',
    cap4: 'AI 對話',
    cap5: '模型設定管理',
    cap6: 'RAG 系統設定',
    cap7: '智慧體編排畫布',
    cap8: '智慧體編排中的 Agent 列表',
    alt1: '截圖：知識庫管理介面',
    alt2: '截圖：知識庫內文件列表',
    alt3: '截圖：文件檢索介面',
    alt4: '截圖：AI 對話介面',
    alt5: '截圖：模型設定管理',
    alt6: '截圖：RAG 系統設定',
    alt7: '截圖：智慧體編排畫布',
    alt8: '截圖：智慧體編排 Agent 列表',
  },
};

/** 其余语种：在英文骨架上覆盖翻译（与 en 结构一致） */
const ES = deepMerge(EN, {
  metaHub: {
    title: 'Documentación — LightningRAG',
    description:
      'Documentación de LightningRAG: inicio rápido, desarrollo local, bases de conocimiento, modelos, chat y agentes.',
  },
  metaInit: {
    title: 'Inicio rápido — LightningRAG',
    description:
      'Configure Node.js y Go, ejecute la API Gin y el admin Vue 3, y abra Swagger.',
  },
  metaUse: {
    title: 'Funciones RAG — LightningRAG',
    description:
      'Uso de bases de conocimiento, proveedores de modelos, chat RAG con citas y flujos de agentes.',
  },
  metaPreview: {
    title: 'Interfaz del producto — LightningRAG',
    description:
      'Capturas del panel de administración: bases de conocimiento, recuperación, chat, modelos, RAG y agentes.',
  },
  a11y: {
    skipMain: 'Ir al contenido principal',
    navMain: 'Navegación principal',
    brandHome: 'Inicio de LightningRAG',
  },
  nav: {
    features: 'Características',
    advantages: 'Ventajas',
    articles: 'Artículos',
    docs: 'Documentación',
  },
  breadcrumb: {
    hub: 'Documentación',
    init: 'Inicio rápido',
    using: 'RAG y agentes',
    preview: 'Interfaz del producto',
  },
  ui: { langAria: 'Idioma' },
  footer: {
    copy: '© Comunidad de código abierto LightningRAG. Todos los derechos reservados.',
    home: 'Inicio',
    docs: 'Documentación',
    articles: 'Artículos',
    sitemap: 'Mapa del sitio',
    githubRepo: 'Repositorio en GitHub',
  },
  docHub: {
    title: 'Documentación',
    lead:
      'Aprenda a ejecutar LightningRAG en local y use bases de conocimiento, modelos, chat con recuperación y agentes en la consola.',
    c1t: 'Inicio rápido',
    c1p: 'Requisitos, clonar, servidor Go y frontend Vue, Swagger y Docker.',
    c1a: 'Abrir inicio rápido →',
    c2t: 'RAG y agentes',
    c2p: 'Bases de conocimiento, incrustaciones, chat con citas y orquestación visual.',
    c2a: 'Abrir guía de uso →',
    c3t: 'Código y licencia',
    c3p: 'Apache 2.0. Consulte el README para términos comerciales y contribuciones.',
    c3a: 'Ver en GitHub →',
    c4t: 'Interfaz del producto',
    c4p:
      'Capturas del admin: bases de conocimiento, listas de documentos, recuperación, chat IA, modelos, RAG y agentes.',
    c4a: 'Ver interfaz del producto →',
  },
  pageInit: {
    title: 'Inicio rápido',
    lead:
      'LightningRAG es un proyecto full stack Go (Gin) + Vue 3. Ejecute la API y el admin localmente y explore Swagger.',
    prereqH: 'Requisitos',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git; IDE recomendado (GoLand o VS Code)',
    serverH: 'Backend (server)',
    s1:
      'Abra la carpeta server como raíz del proyecto en el IDE (no la raíz del repositorio). Luego:',
    sNote:
      'La API suele escuchar en el puerto 8888. Ajuste base de datos y Redis en la configuración de server/.',
    webH: 'Frontend (web)',
    w1:
      'En web/, instale dependencias y arranque el servidor de desarrollo. Configure la URL base de la API si es necesario.',
    swaggerH: 'Swagger',
    sw1:
      'Instale swag, ejecute swag init en server/, reinicie y abra /swagger/index.html.',
    dockerH: 'Docker y Compose',
    d1:
      'Vea deploy/docker-compose en el repositorio. Ajuste perfiles y .env para el almacén vectorial.',
    moreH: 'Siguientes pasos',
    m1:
      'Lea el README del repositorio para menús Casbin, generación de código y módulo RAG. Uso comercial: Apache 2.0.',
  },
  pageUse: {
    title: 'Uso de funciones RAG',
    lead:
      'La consola de administración integra bases de conocimiento, modelos, chat con recuperación y diseñador de agentes.',
    kbH: 'Bases de conocimiento',
    kb1:
      'Cree una base, suba documentos, elija opciones de fragmentación y espere la indexación.',
    kb2:
      'Asocie modelos de embedding y rerank o use valores por defecto del sistema.',
    modelH: 'Modelos y proveedores',
    mo1:
      'Configure LLM, embedding y rerank en ajustes del sistema. Se admiten muchos endpoints compatibles con OpenAI y runtimes locales.',
    chatH: 'Chat RAG',
    ch1:
      'Inicie un chat, seleccione bases de conocimiento y revise las citas enlazadas a fragmentos recuperados.',
    agentH: 'Agentes y flujos',
    ag1:
      'Use el lienzo: Begin, Retrieval, LLM, Message, ramas, HTTP, etc. Guarde versiones y ejecute con consultas.',
    securityH: 'Usuarios y permisos',
    sec1:
      'JWT, rutas dinámicas y Casbin alinean menús y API con roles para despliegues privados.',
  },
});

const FR = deepMerge(EN, {
  metaHub: {
    title: 'Documentation — LightningRAG',
    description:
      'Documentation LightningRAG : démarrage rapide, développement local, bases de connaissances, modèles, chat et agents.',
  },
  metaInit: {
    title: 'Démarrage rapide — LightningRAG',
    description:
      'Configurer Node.js et Go, lancer l’API Gin et l’admin Vue 3, ouvrir Swagger.',
  },
  metaUse: {
    title: 'Fonctions RAG — LightningRAG',
    description:
      'Bases de connaissances, fournisseurs de modèles, chat RAG avec citations et workflows agents.',
  },
  metaPreview: {
    title: 'Interface produit — LightningRAG',
    description:
      'Captures de la console d’administration : bases de connaissances, recherche, chat, modèles, RAG et agents.',
  },
  a11y: {
    skipMain: 'Aller au contenu principal',
    navMain: 'Navigation principale',
    brandHome: 'Accueil LightningRAG',
  },
  nav: {
    features: 'Fonctionnalités',
    advantages: 'Avantages',
    articles: 'Articles',
    docs: 'Documentation',
  },
  breadcrumb: {
    hub: 'Documentation',
    init: 'Démarrage rapide',
    using: 'RAG et agents',
    preview: 'Interface produit',
  },
  ui: { langAria: 'Langue' },
  footer: {
    copy: '© Communauté open source LightningRAG. Tous droits réservés.',
    home: 'Accueil',
    docs: 'Documentation',
    articles: 'Articles',
    sitemap: 'Plan du site',
    githubRepo: 'Dépôt GitHub',
  },
  docHub: {
    title: 'Documentation',
    lead:
      'Exécutez LightningRAG en local, puis utilisez bases de connaissances, modèles, chat de récupération et agents.',
    c1t: 'Démarrage rapide',
    c1p: 'Prérequis, cloner, serveur Go et Vue, Swagger et Docker.',
    c1a: 'Ouvrir le démarrage rapide →',
    c2t: 'RAG et agents',
    c2p: 'Bases de connaissances, embeddings, chat avec citations, orchestration visuelle.',
    c2a: 'Ouvrir le guide d’utilisation →',
    c3t: 'Source et licence',
    c3p: 'Apache 2.0. Voir le README pour l’usage commercial et les contributions.',
    c3a: 'Voir sur GitHub →',
    c4t: 'Interface produit',
    c4p:
      'Captures de l’admin : bases de connaissances, listes de documents, recherche, chat IA, modèles, RAG et agents.',
    c4a: 'Voir l’interface produit →',
  },
  pageInit: {
    title: 'Démarrage rapide',
    lead:
      'LightningRAG est une stack Go (Gin) + Vue 3. Lancez l’API et l’admin localement, explorez Swagger.',
    prereqH: 'Prérequis',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git ; IDE recommandé',
    serverH: 'Backend (server)',
    s1:
      'Ouvrez le dossier server comme racine du projet dans l’IDE (pas la racine du dépôt). Puis :',
    sNote:
      'L’API écoute par défaut sur le port 8888. Configurez BDD et Redis dans server/.',
    webH: 'Frontend (web)',
    w1:
      'Dans web/, installez les dépendances et lancez le serveur de dev. Ajustez l’URL de l’API si besoin.',
    swaggerH: 'Swagger',
    sw1:
      'Installez swag, exécutez swag init dans server/, redémarrez, ouvrez /swagger/index.html.',
    dockerH: 'Docker et Compose',
    d1:
      'Voir deploy/docker-compose. Ajustez profils et .env pour le magasin vectoriel.',
    moreH: 'Étapes suivantes',
    m1:
      'Lisez le README pour Casbin, génération de code et module RAG. Usage commercial : Apache 2.0.',
  },
  pageUse: {
    title: 'Utilisation du RAG',
    lead:
      'La console admin regroupe bases de connaissances, modèles, chat avec récupération et concepteur d’agents.',
    kbH: 'Bases de connaissances',
    kb1:
      'Créez une base, importez des documents, choisissez le découpage et attendez l’indexation.',
    kb2:
      'Liez embedding et rerank optionnel ou utilisez les paramètres système.',
    modelH: 'Modèles et fournisseurs',
    mo1:
      'Configurez LLM, embedding, rerank. Nombreux endpoints compatibles OpenAI et runtimes locaux.',
    chatH: 'Conversation RAG',
    ch1:
      'Sélectionnez des bases, posez des questions, consultez les citations vers les extraits.',
    agentH: 'Agents et flux',
    ag1:
      'Canvas : Begin, Retrieval, LLM, Message, branches, HTTP. Enregistrez des versions et exécutez.',
    securityH: 'Utilisateurs et droits',
    sec1:
      'JWT, routes dynamiques et Casbin alignent menus et API sur les rôles.',
  },
});

const DE = deepMerge(EN, {
  metaHub: {
    title: 'Dokumentation — LightningRAG',
    description:
      'LightningRAG-Dokumentation: Schnellstart, lokale Entwicklung, Wissensbasen, Modelle, Chat und Agenten.',
  },
  metaInit: {
    title: 'Schnellstart — LightningRAG',
    description:
      'Node.js und Go einrichten, Gin-API und Vue-3-Admin starten, Swagger öffnen.',
  },
  metaUse: {
    title: 'RAG-Funktionen — LightningRAG',
    description:
      'Wissensbasen, Modellanbieter, RAG-Chat mit Zitaten und Agenten-Workflows.',
  },
  metaPreview: {
    title: 'Produkt-Oberfläche — LightningRAG',
    description:
      'Screenshots der Admin-Konsole: Wissensbasen, Suche, Chat, Modelle, RAG und Agenten.',
  },
  a11y: {
    skipMain: 'Zum Hauptinhalt springen',
    navMain: 'Hauptnavigation',
    brandHome: 'LightningRAG Startseite',
  },
  nav: {
    features: 'Funktionen',
    advantages: 'Vorteile',
    articles: 'Artikel',
    docs: 'Dokumentation',
  },
  breadcrumb: {
    hub: 'Dokumentation',
    init: 'Schnellstart',
    using: 'RAG & Agenten',
    preview: 'Produkt-Oberfläche',
  },
  ui: { langAria: 'Sprache' },
  footer: {
    copy: '© LightningRAG Open-Source-Community. Alle Rechte vorbehalten.',
    home: 'Startseite',
    docs: 'Dokumentation',
    articles: 'Artikel',
    sitemap: 'Sitemap',
    githubRepo: 'GitHub-Repository',
  },
  docHub: {
    title: 'Dokumentation',
    lead:
      'LightningRAG lokal ausführen und Wissensbasen, Modelle, Retrieval-Chat und Agenten in der Konsole nutzen.',
    c1t: 'Schnellstart',
    c1p: 'Voraussetzungen, Klonen, Go-Server und Vue-Frontend, Swagger und Docker.',
    c1a: 'Schnellstart öffnen →',
    c2t: 'RAG und Agenten',
    c2p: 'Wissensbasen, Embeddings, Chat mit Zitaten, visuelle Orchestrierung.',
    c2a: 'Nutzungsleitfaden öffnen →',
    c3t: 'Quellcode und Lizenz',
    c3p: 'Apache 2.0. README für kommerzielle Nutzung und Beiträge.',
    c3a: 'Auf GitHub ansehen →',
    c4t: 'Produkt-Oberfläche',
    c4p:
      'Screenshots der Konsole: Wissensbasen, Dokumentlisten, Suche, KI-Chat, Modelle, RAG und Agenten.',
    c4a: 'Produkt-Oberfläche ansehen →',
  },
  pageInit: {
    title: 'Schnellstart',
    lead:
      'LightningRAG ist Go (Gin) + Vue 3 Full-Stack. API und Admin lokal starten, Swagger nutzen.',
    prereqH: 'Voraussetzungen',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git; IDE empfohlen',
    serverH: 'Backend (server)',
    s1:
      'Ordner server als Projektroot in der IDE öffnen (nicht das Repo-Root). Dann:',
    sNote:
      'API standardmäßig Port 8888. Datenbank und Redis in server/ konfigurieren.',
    webH: 'Frontend (web)',
    w1:
      'In web/ Abhängigkeiten installieren und Dev-Server starten. API-URL bei Bedarf anpassen.',
    swaggerH: 'Swagger',
    sw1: 'swag installieren, in server/ swag init, neu starten, /swagger/index.html.',
    dockerH: 'Docker & Compose',
    d1: 'Siehe deploy/docker-compose im Repository. Profile und .env anpassen.',
    moreH: 'Nächste Schritte',
    m1:
      'README für Casbin, Codegenerierung und RAG-Modul lesen. Kommerziell: Apache 2.0.',
  },
  pageUse: {
    title: 'RAG-Funktionen nutzen',
    lead:
      'Admin-Konsole: Wissensbasen, Modelle, Retrieval-Chat und Agenten-Designer.',
    kbH: 'Wissensbasen',
    kb1: 'Basis anlegen, Dokumente hochladen, Chunking wählen, Indexierung abwarten.',
    kb2: 'Embedding und optional Rerank zuweisen oder Systemdefaults nutzen.',
    modelH: 'Modelle & Anbieter',
    mo1:
      'LLM, Embedding, Rerank in den Einstellungen. Viele OpenAI-kompatible und lokale Laufzeiten.',
    chatH: 'RAG-Gespräch',
    ch1:
      'Wissensbasen wählen, Fragen stellen, Zitate zu abgerufenen Chunks prüfen.',
    agentH: 'Agenten & Abläufe',
    ag1:
      'Canvas: Begin, Retrieval, LLM, Message, Verzweigungen, HTTP. Versionen speichern und ausführen.',
    securityH: 'Nutzer & Rechte',
    sec1: 'JWT, dynamische Routen und Casbin für Menüs und APIs nach Rollen.',
  },
});

const JA = deepMerge(EN, {
  metaHub: {
    title: 'ドキュメント — LightningRAG',
    description:
      'LightningRAG のドキュメント：クイックスタート、ローカル開発、ナレッジベース、モデル、チャット、エージェント。',
  },
  metaInit: {
    title: 'クイックスタート — LightningRAG',
    description:
      'Node.js と Go を用意し、Gin API と Vue 3 管理画面を起動、Swagger で API を確認。',
  },
  metaUse: {
    title: 'RAG 機能の使い方 — LightningRAG',
    description:
      'ナレッジベース、モデルプロバイダー、引用付き RAG チャット、エージェントワークフロー。',
  },
  metaPreview: {
    title: '製品画面 — LightningRAG',
    description:
      '管理コンソールのスクリーンショット：ナレッジベース、検索、チャット、モデル、RAG、エージェント。',
  },
  a11y: {
    skipMain: 'メインコンテンツへスキップ',
    navMain: 'メインナビゲーション',
    brandHome: 'LightningRAG ホーム',
  },
  nav: {
    features: '特徴',
    advantages: '利点',
    articles: '記事',
    docs: 'ドキュメント',
  },
  breadcrumb: {
    hub: 'ドキュメント',
    init: 'クイックスタート',
    using: 'RAG とエージェント',
    preview: '製品画面',
  },
  ui: { langAria: '言語' },
  footer: {
    copy: '© LightningRAG オープンソースコミュニティ。無断転載を禁じます。',
    home: 'ホーム',
    docs: 'ドキュメント',
    articles: '記事',
    sitemap: 'サイトマップ',
    githubRepo: 'GitHub リポジトリ',
  },
  docHub: {
    title: 'ドキュメント',
    lead:
      'ローカルで LightningRAG を動かし、管理コンソールでナレッジベース、モデル、検索チャット、エージェントを利用します。',
    c1t: 'クイックスタート',
    c1p: '前提条件、クローン、Go サーバーと Vue、Swagger／Docker。',
    c1a: 'クイックスタートへ →',
    c2t: 'RAG とエージェント',
    c2p: 'ナレッジベース、埋め込み、引用付きチャット、ビジュアル編排。',
    c2a: '利用ガイドへ →',
    c3t: 'ソースとライセンス',
    c3p: 'Apache 2.0。商用・貢献は README を参照。',
    c3a: 'GitHub で見る →',
    c4t: '製品画面',
    c4p:
      '管理コンソールのスクリーンショット：ナレッジベース、文書一覧、検索、AI チャット、モデル、RAG、エージェント。',
    c4a: '製品画面を見る →',
  },
  pageInit: {
    title: 'クイックスタート',
    lead:
      'LightningRAG は Go（Gin）+ Vue 3 のフルスタックです。API と管理 UI を起動し Swagger で確認します。',
    prereqH: '前提条件',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git、推奨 IDE',
    serverH: 'バックエンド（server）',
    s1:
      'IDE では server フォルダをプロジェクトルートに（リポジトリルートではない）。次を実行：',
    sNote:
      'API は既定でポート 8888。DB・Redis は server 配下の設定で変更。',
    webH: 'フロントエンド（web）',
    w1:
      'web で依存関係をインストールし開発サーバーを起動。必要なら API の URL を設定。',
    swaggerH: 'Swagger',
    sw1: 'swag を入れ server で swag init、再起動し /swagger/index.html を開く。',
    dockerH: 'Docker / Compose',
    d1: 'リポジトリの deploy/docker-compose を参照。ベクトル DB に合わせ profile と .env を調整。',
    moreH: '次のステップ',
    m1:
      'README で Casbin、コード生成、RAG モジュールを確認。商用は Apache 2.0 に従ってください。',
  },
  pageUse: {
    title: 'RAG 機能の利用',
    lead:
      '管理コンソールにナレッジベース、モデル設定、検索チャット、エージェント設計がまとまっています。',
    kbH: 'ナレッジベース',
    kb1: 'ベースを作成しドキュメントをアップロード、チャンク設定を選びインデックス完了を待つ。',
    kb2: '埋め込みと任意のリランクを紐付け、またはシステム既定を使用。',
    modelH: 'モデルとプロバイダー',
    mo1:
      'システム設定で LLM・埋め込み・リランクを設定。OpenAI 互換やローカル実行を多数サポート。',
    chatH: '会話型 RAG',
    ch1: 'チャットでナレッジベースを選び質問。取得チャンクに対応する引用を確認。',
    agentH: 'エージェントとワークフロー',
    ag1:
      'キャンバスで Begin、Retrieval、LLM、Message、分岐、HTTP などを配置。版を保存して実行。',
    securityH: 'ユーザーと権限',
    sec1: 'JWT・動的ルート・Casbin でメニューと API をロールに合わせる。',
  },
});

const KO = deepMerge(EN, {
  metaHub: {
    title: '문서 — LightningRAG',
    description:
      'LightningRAG 문서: 빠른 시작, 로컬 개발, 지식 베이스, 모델, 채팅, 에이전트.',
  },
  metaInit: {
    title: '빠른 시작 — LightningRAG',
    description:
      'Node.js와 Go를 준비하고 Gin API와 Vue 3 관리 화면을 실행한 뒤 Swagger로 API를 확인합니다.',
  },
  metaUse: {
    title: 'RAG 기능 사용 — LightningRAG',
    description:
      '지식 베이스, 모델 공급자, 인용이 있는 RAG 채팅, 에이전트 워크플로.',
  },
  metaPreview: {
    title: '제품 화면 — LightningRAG',
    description:
      '관리 콘솔 스크린샷: 지식 베이스, 검색, 채팅, 모델, RAG, 에이전트.',
  },
  a11y: {
    skipMain: '본문으로 건너뛰기',
    navMain: '주 내비게이션',
    brandHome: 'LightningRAG 홈',
  },
  nav: {
    features: '특징',
    advantages: '장점',
    articles: '글',
    docs: '문서',
  },
  breadcrumb: {
    hub: '문서',
    init: '빠른 시작',
    using: 'RAG 및 에이전트',
    preview: '제품 화면',
  },
  ui: { langAria: '언어' },
  footer: {
    copy: '© LightningRAG 오픈소스 커뮤니티. 무단 복제를 금합니다.',
    home: '홈',
    docs: '문서',
    articles: '글',
    sitemap: '사이트맵',
    githubRepo: 'GitHub 저장소',
  },
  docHub: {
    title: '문서',
    lead:
      '로컬에서 LightningRAG를 실행하고 콘솔에서 지식 베이스, 모델, 검색 채팅, 에이전트를 사용합니다.',
    c1t: '빠른 시작',
    c1p: '요구 사항, 클론, Go 서버와 Vue, Swagger/Docker.',
    c1a: '빠른 시작 열기 →',
    c2t: 'RAG 및 에이전트',
    c2p: '지식 베이스, 임베딩, 인용 채팅, 시각적 오케스트레이션.',
    c2a: '사용 가이드 열기 →',
    c3t: '소스 및 라이선스',
    c3p: 'Apache 2.0. 상업적 이용·기여는 README를 참고하세요.',
    c3a: 'GitHub에서 보기 →',
    c4t: '제품 화면',
    c4p:
      '관리 콘솔 스크린샷: 지식 베이스, 문서 목록, 검색, AI 채팅, 모델, RAG, 에이전트.',
    c4a: '제품 화면 보기 →',
  },
  pageInit: {
    title: '빠른 시작',
    lead:
      'LightningRAG는 Go(Gin)+Vue 3 풀스택입니다. API와 관리 UI를 띄운 뒤 Swagger를 확인합니다.',
    prereqH: '요구 사항',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git, 권장 IDE',
    serverH: '백엔드(server)',
    s1:
      'IDE에서 server 폴더를 프로젝트 루트로 엽니다(저장소 루트 아님). 다음 실행:',
    sNote:
      'API 기본 포트 8888. DB·Redis는 server 설정에서 변경.',
    webH: '프론트엔드(web)',
    w1:
      'web에서 의존성 설치 후 개발 서버 실행. 필요 시 API URL 설정.',
    swaggerH: 'Swagger',
    sw1: 'swag 설치 후 server에서 swag init, 재시작 후 /swagger/index.html.',
    dockerH: 'Docker 및 Compose',
    d1: '저장소의 deploy/docker-compose 참고. 벡터 저장소에 맞게 profile·.env 조정.',
    moreH: '다음 단계',
    m1:
      'README에서 Casbin, 코드 생성, RAG 모듈 확인. 상업적 이용은 Apache 2.0 준수.',
  },
  pageUse: {
    title: 'RAG 기능 사용',
    lead:
      '관리 콘솔에 지식 베이스, 모델, 검색 채팅, 에이전트 디자이너가 통합되어 있습니다.',
    kbH: '지식 베이스',
    kb1: '베이스를 만들고 문서를 업로드, 청킹 옵션을 선택하고 인덱싱을 기다립니다.',
    kb2: '임베딩과 선택적 리랭크를 연결하거나 시스템 기본값을 사용합니다.',
    modelH: '모델 및 공급자',
    mo1:
      '시스템 설정에서 LLM·임베딩·리랭크 구성. OpenAI 호환 및 로컬 런타임 다수 지원.',
    chatH: '대화형 RAG',
    ch1: '채팅에서 지식 베이스를 선택하고 질문. 검색된 청크와 연결된 인용을 확인합니다.',
    agentH: '에이전트 및 워크플로',
    ag1:
      '캔버스에서 Begin, Retrieval, LLM, Message, 분기, HTTP 등 구성. 버전 저장 후 실행.',
    securityH: '사용자 및 권한',
    sec1: 'JWT, 동적 라우트, Casbin으로 메뉴·API를 역할에 맞춥니다.',
  },
});

const PT = deepMerge(EN, {
  metaHub: {
    title: 'Documentação — LightningRAG',
    description:
      'Documentação LightningRAG: início rápido, desenvolvimento local, bases de conhecimento, modelos, chat e agentes.',
  },
  metaInit: {
    title: 'Início rápido — LightningRAG',
    description:
      'Configure Node.js e Go, execute a API Gin e o admin Vue 3, abra o Swagger.',
  },
  metaUse: {
    title: 'Recursos RAG — LightningRAG',
    description:
      'Bases de conhecimento, provedores de modelos, chat RAG com citações e fluxos de agentes.',
  },
  metaPreview: {
    title: 'Interface do produto — LightningRAG',
    description:
      'Capturas do console de administração: bases de conhecimento, busca, chat, modelos, RAG e agentes.',
  },
  a11y: {
    skipMain: 'Ir para o conteúdo principal',
    navMain: 'Navegação principal',
    brandHome: 'Página inicial LightningRAG',
  },
  nav: {
    features: 'Recursos',
    advantages: 'Vantagens',
    articles: 'Artigos',
    docs: 'Documentação',
  },
  breadcrumb: {
    hub: 'Documentação',
    init: 'Início rápido',
    using: 'RAG e agentes',
    preview: 'Interface do produto',
  },
  ui: { langAria: 'Idioma' },
  footer: {
    copy: '© Comunidade open source LightningRAG. Todos os direitos reservados.',
    home: 'Início',
    docs: 'Documentação',
    articles: 'Artigos',
    sitemap: 'Mapa do site',
    githubRepo: 'Repositório GitHub',
  },
  docHub: {
    title: 'Documentação',
    lead:
      'Execute o LightningRAG localmente e use bases de conhecimento, modelos, chat com recuperação e agentes no console.',
    c1t: 'Início rápido',
    c1p: 'Requisitos, clonar, servidor Go e Vue, Swagger e Docker.',
    c1a: 'Abrir início rápido →',
    c2t: 'RAG e agentes',
    c2p: 'Bases de conhecimento, embeddings, chat com citações, orquestração visual.',
    c2a: 'Abrir guia de uso →',
    c3t: 'Código e licença',
    c3p: 'Apache 2.0. Veja o README para uso comercial e contribuições.',
    c3a: 'Ver no GitHub →',
    c4t: 'Interface do produto',
    c4p:
      'Capturas do admin: bases de conhecimento, listas de documentos, busca, chat IA, modelos, RAG e agentes.',
    c4a: 'Ver interface do produto →',
  },
  pageInit: {
    title: 'Início rápido',
    lead:
      'LightningRAG é full stack Go (Gin) + Vue 3. Execute a API e o admin localmente e explore o Swagger.',
    prereqH: 'Requisitos',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git; IDE recomendada',
    serverH: 'Backend (server)',
    s1:
      'Abra a pasta server como raiz do projeto na IDE (não a raiz do repositório). Depois:',
    sNote:
      'A API escuta na porta 8888 por padrão. Configure BD e Redis em server/.',
    webH: 'Frontend (web)',
    w1:
      'Em web/, instale dependências e inicie o servidor de desenvolvimento. Ajuste a URL da API se necessário.',
    swaggerH: 'Swagger',
    sw1:
      'Instale swag, execute swag init em server/, reinicie, abra /swagger/index.html.',
    dockerH: 'Docker e Compose',
    d1:
      'Veja deploy/docker-compose no repositório. Ajuste perfis e .env para o armazenamento vetorial.',
    moreH: 'Próximos passos',
    m1:
      'Leia o README para Casbin, geração de código e módulo RAG. Uso comercial: Apache 2.0.',
  },
  pageUse: {
    title: 'Uso dos recursos RAG',
    lead:
      'O console admin integra bases de conhecimento, modelos, chat com recuperação e designer de agentes.',
    kbH: 'Bases de conhecimento',
    kb1:
      'Crie uma base, envie documentos, escolha opções de fragmentação e aguarde a indexação.',
    kb2:
      'Associe modelos de embedding e rerank opcional ou use padrões do sistema.',
    modelH: 'Modelos e provedores',
    mo1:
      'Configure LLM, embedding e rerank nas definições. Muitos endpoints compatíveis com OpenAI e runtimes locais.',
    chatH: 'Chat RAG',
    ch1:
      'Inicie um chat, selecione bases, faça perguntas e veja citações ligadas aos trechos recuperados.',
    agentH: 'Agentes e fluxos',
    ag1:
      'Use a tela: Begin, Retrieval, LLM, Message, ramificações, HTTP. Salve versões e execute.',
    securityH: 'Usuários e permissões',
    sec1:
      'JWT, rotas dinâmicas e Casbin alinham menus e APIs às funções.',
  },
});

const RU = deepMerge(EN, {
  metaHub: {
    title: 'Документация — LightningRAG',
    description:
      'Документация LightningRAG: быстрый старт, локальная разработка, базы знаний, модели, чат и агенты.',
  },
  metaInit: {
    title: 'Быстрый старт — LightningRAG',
    description:
      'Настройте Node.js и Go, запустите API на Gin и админку Vue 3, откройте Swagger.',
  },
  metaUse: {
    title: 'Функции RAG — LightningRAG',
    description:
      'Базы знаний, поставщики моделей, RAG-чат с цитатами и сценарии агентов.',
  },
  a11y: {
    skipMain: 'Перейти к основному содержимому',
    navMain: 'Основная навигация',
    brandHome: 'Главная LightningRAG',
  },
  nav: {
    features: 'Возможности',
    advantages: 'Преимущества',
    articles: 'Статьи',
    docs: 'Документация',
  },
  breadcrumb: {
    hub: 'Документация',
    init: 'Быстрый старт',
    using: 'RAG и агенты',
    preview: 'Интерфейс продукта',
  },
  ui: { langAria: 'Язык' },
  footer: {
    copy: '© Сообщество LightningRAG с открытым исходным кодом. Все права защищены.',
    home: 'Главная',
    docs: 'Документация',
    articles: 'Статьи',
    sitemap: 'Карта сайта',
    githubRepo: 'Репозиторий GitHub',
  },
  docHub: {
    title: 'Документация',
    lead:
      'Запустите LightningRAG локально и используйте базы знаний, модели, чат с поиском и агентов в консоли.',
    c1t: 'Быстрый старт',
    c1p: 'Требования, клонирование, сервер Go и Vue, Swagger и Docker.',
    c1a: 'Открыть быстрый старт →',
    c2t: 'RAG и агенты',
    c2p: 'Базы знаний, эмбеддинги, чат с цитатами, визуальная оркестрация.',
    c2a: 'Открыть руководство →',
    c3t: 'Исходный код и лицензия',
    c3p: 'Apache 2.0. Коммерческое использование и вклад — в README.',
    c3a: 'На GitHub →',
  },
  pageInit: {
    title: 'Быстрый старт',
    lead:
      'LightningRAG — полный стек Go (Gin) + Vue 3. Запустите API и админку локально, изучите Swagger.',
    prereqH: 'Требования',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git; рекомендуется IDE',
    serverH: 'Бэкенд (server)',
    s1:
      'Откройте папку server как корень проекта в IDE (не корень репозитория). Затем:',
    sNote:
      'API по умолчанию на порту 8888. Настройте БД и Redis в конфигурации server/.',
    webH: 'Фронтенд (web)',
    w1:
      'В web/ установите зависимости и запустите dev-сервер. При необходимости укажите URL API.',
    swaggerH: 'Swagger',
    sw1:
      'Установите swag, выполните swag init в server/, перезапустите, откройте /swagger/index.html.',
    dockerH: 'Docker и Compose',
    d1:
      'См. deploy/docker-compose в репозитории. Настройте профили и .env под векторное хранилище.',
    moreH: 'Дальше',
    m1:
      'README: Casbin, генерация кода, модуль RAG. Коммерция: Apache 2.0.',
  },
  pageUse: {
    title: 'Использование RAG',
    lead:
      'Админ-консоль объединяет базы знаний, модели, чат с поиском и конструктор агентов.',
    kbH: 'Базы знаний',
    kb1:
      'Создайте базу, загрузите документы, выберите разбиение на фрагменты, дождитесь индексации.',
    kb2:
      'Привяжите эмбеддинги и опциональный rerank или используйте системные значения по умолчанию.',
    modelH: 'Модели и провайдеры',
    mo1:
      'Настройте LLM, эмбеддинги, rerank в системных настройках. Поддерживаются OpenAI-совместимые и локальные среды.',
    chatH: 'Диалоговый RAG',
    ch1:
      'Выберите базы знаний, задавайте вопросы, смотрите цитаты к извлечённым фрагментам.',
    agentH: 'Агенты и сценарии',
    ag1:
      'Холст: Begin, Retrieval, LLM, Message, ветвления, HTTP. Сохраняйте версии и запускайте.',
    securityH: 'Пользователи и права',
    sec1:
      'JWT, динамические маршруты и Casbin связывают меню и API с ролями.',
  },
});

const IT = deepMerge(EN, {
  metaHub: {
    title: 'Documentazione — LightningRAG',
    description:
      'Documentazione LightningRAG: avvio rapido, sviluppo locale, basi di conoscenza, modelli, chat e agenti.',
  },
  metaInit: {
    title: 'Avvio rapido — LightningRAG',
    description:
      'Configura Node.js e Go, avvia API Gin e admin Vue 3, apri Swagger.',
  },
  metaUse: {
    title: 'Funzionalità RAG — LightningRAG',
    description:
      'Basi di conoscenza, provider di modelli, chat RAG con citazioni e flussi agente.',
  },
  a11y: {
    skipMain: 'Vai al contenuto principale',
    navMain: 'Navigazione principale',
    brandHome: 'Home LightningRAG',
  },
  nav: {
    features: 'Funzionalità',
    advantages: 'Vantaggi',
    articles: 'Articoli',
    docs: 'Documentazione',
  },
  breadcrumb: {
    hub: 'Documentazione',
    init: 'Avvio rapido',
    using: 'RAG e agenti',
    preview: 'Interfaccia prodotto',
  },
  ui: { langAria: 'Lingua' },
  footer: {
    copy: '© Comunità open source LightningRAG. Tutti i diritti riservati.',
    home: 'Home',
    docs: 'Documentazione',
    articles: 'Articoli',
    sitemap: 'Mappa del sito',
    githubRepo: 'Repository GitHub',
  },
  docHub: {
    title: 'Documentazione',
    lead:
      'Esegui LightningRAG in locale e usa basi di conoscenza, modelli, chat con recupero e agenti nella console.',
    c1t: 'Avvio rapido',
    c1p: 'Prerequisiti, clone, server Go e Vue, Swagger e Docker.',
    c1a: 'Apri avvio rapido →',
    c2t: 'RAG e agenti',
    c2p: 'Basi di conoscenza, embedding, chat con citazioni, orchestrazione visiva.',
    c2a: 'Apri guida uso →',
    c3t: 'Codice e licenza',
    c3p: 'Apache 2.0. README per uso commerciale e contributi.',
    c3a: 'Vedi su GitHub →',
  },
  pageInit: {
    title: 'Avvio rapido',
    lead:
      'LightningRAG è stack completo Go (Gin) + Vue 3. Avvia API e admin in locale e usa Swagger.',
    prereqH: 'Prerequisiti',
    p1: 'Node.js > v18.16.0',
    p2: 'Go >= 1.22',
    p3: 'Git; IDE consigliata',
    serverH: 'Backend (server)',
    s1:
      'Apri la cartella server come radice del progetto nell’IDE (non la radice del repo). Poi:',
    sNote:
      'L’API ascolta sulla porta 8888 di default. Configura DB e Redis in server/.',
    webH: 'Frontend (web)',
    w1:
      'In web/ installa le dipendenze e avvia il dev server. Imposta l’URL dell’API se serve.',
    swaggerH: 'Swagger',
    sw1:
      'Installa swag, esegui swag init in server/, riavvia, apri /swagger/index.html.',
    dockerH: 'Docker e Compose',
    d1:
      'Vedi deploy/docker-compose nel repository. Adatta profili e .env allo store vettoriale.',
    moreH: 'Passi successivi',
    m1:
      'Leggi il README per Casbin, generazione codice e modulo RAG. Uso commerciale: Apache 2.0.',
  },
  pageUse: {
    title: 'Uso delle funzionalità RAG',
    lead:
      'La console admin integra basi di conoscenza, modelli, chat con recupero e designer di agenti.',
    kbH: 'Basi di conoscenza',
    kb1:
      'Crea una base, carica documenti, scegli il chunking e attendi l’indicizzazione.',
    kb2:
      'Collega embedding e rerank opzionale o usa i default di sistema.',
    modelH: 'Modelli e provider',
    mo1:
      'Configura LLM, embedding e rerank nelle impostazioni. Supportati molti endpoint compatibili OpenAI e runtime locali.',
    chatH: 'Chat RAG',
    ch1:
      'Avvia una chat, seleziona le basi, poni domande e controlla le citazioni ai chunk recuperati.',
    agentH: 'Agenti e flussi',
    ag1:
      'Tela: Begin, Retrieval, LLM, Message, diramazioni, HTTP. Salva versioni ed esegui.',
    securityH: 'Utenti e permessi',
    sec1:
      'JWT, route dinamiche e Casbin allineano menu e API ai ruoli.',
  },
});

export default {
  en: EN,
  'zh-CN': deepMerge(EN, ZH_CN),
  'zh-TW': deepMerge(EN, ZH_TW),
  es: ES,
  fr: FR,
  de: DE,
  ja: JA,
  ko: KO,
  pt: PT,
  ru: RU,
  it: IT,
};
