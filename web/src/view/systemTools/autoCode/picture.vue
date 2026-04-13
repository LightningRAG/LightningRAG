<template>
  <div>
    <warning-bar
        href="https://plugin.LightningRAG.com/license"
        :title="$t('tools.aiPicture.licenseTitle')"
    />
    <div class="lrag-search-box">
      <div class="text-xl mb-2 text-gray-600">
        {{ $t('tools.aiPicture.subtitle') }}
      </div>
      
      <!-- 选项模式 -->
      <div class="mb-4">
        <div class="mb-3">
          <div class="text-base font-medium mb-2">{{ $t('tools.aiPicture.secPageUse') }}</div>
          <el-radio-group v-model="pageType" class="mb-2" @change="handlePageTypeChange">
            <el-radio label="corporate">{{ $t('tools.aiPicture.ptCorp') }}</el-radio>
            <el-radio label="ecommerce">{{ $t('tools.aiPicture.ptEcom') }}</el-radio>
            <el-radio label="blog">{{ $t('tools.aiPicture.ptBlog') }}</el-radio>
            <el-radio label="product">{{ $t('tools.aiPicture.ptProduct') }}</el-radio>
            <el-radio label="landing">{{ $t('tools.aiPicture.ptLanding') }}</el-radio>
            <el-radio label="other">{{ $t('tools.aiPicture.other') }}</el-radio>
          </el-radio-group>
          <el-input v-if="pageType === 'other'" v-model="pageTypeCustom" :placeholder="$t('tools.aiPicture.phPageUse')" class="w-full" />
        </div>
        
        <div class="mb-3">
          <div class="text-base font-medium mb-2">{{ $t('tools.aiPicture.secBlocks') }}</div>
          <el-checkbox-group v-model="contentBlocks" class="flex flex-wrap gap-2 mb-2">
            <el-checkbox label="banner">{{ $t('tools.aiPicture.cbBanner') }}</el-checkbox>
            <el-checkbox label="product_service">{{ $t('tools.aiPicture.cbProductSvc') }}</el-checkbox>
            <el-checkbox label="features">{{ $t('tools.aiPicture.cbFeatures') }}</el-checkbox>
            <el-checkbox label="cases">{{ $t('tools.aiPicture.cbCases') }}</el-checkbox>
            <el-checkbox label="team">{{ $t('tools.aiPicture.cbTeam') }}</el-checkbox>
            <el-checkbox label="contact">{{ $t('tools.aiPicture.cbContact') }}</el-checkbox>
            <el-checkbox label="news">{{ $t('tools.aiPicture.cbNews') }}</el-checkbox>
            <el-checkbox label="pricing">{{ $t('tools.aiPicture.cbPricing') }}</el-checkbox>
            <el-checkbox label="faq">{{ $t('tools.aiPicture.cbFaq') }}</el-checkbox>
            <el-checkbox label="reviews">{{ $t('tools.aiPicture.cbReviews') }}</el-checkbox>
            <el-checkbox label="stats">{{ $t('tools.aiPicture.cbStats') }}</el-checkbox>
            <el-checkbox label="product_list">{{ $t('tools.aiPicture.cbProductList') }}</el-checkbox>
            <el-checkbox label="product_card">{{ $t('tools.aiPicture.cbProductCard') }}</el-checkbox>
            <el-checkbox label="cart">{{ $t('tools.aiPicture.cbCart') }}</el-checkbox>
            <el-checkbox label="checkout">{{ $t('tools.aiPicture.cbCheckout') }}</el-checkbox>
            <el-checkbox label="order_tracking">{{ $t('tools.aiPicture.cbOrder') }}</el-checkbox>
            <el-checkbox label="category">{{ $t('tools.aiPicture.cbCategory') }}</el-checkbox>
            <el-checkbox label="hot_picks">{{ $t('tools.aiPicture.cbHot') }}</el-checkbox>
            <el-checkbox label="flash_sale">{{ $t('tools.aiPicture.cbSale') }}</el-checkbox>
            <el-checkbox label="other">{{ $t('tools.aiPicture.other') }}</el-checkbox>
          </el-checkbox-group>
          <el-input v-if="contentBlocks.includes('other')" v-model="contentBlocksCustom" :placeholder="$t('tools.aiPicture.phBlocks')" class="w-full" />
        </div>
        
        <div class="mb-3">
          <div class="text-base font-medium mb-2">{{ $t('tools.aiPicture.secStyle') }}</div>
          <el-radio-group v-model="stylePreference" class="mb-2">
            <el-radio label="minimal">{{ $t('tools.aiPicture.stMinimal') }}</el-radio>
            <el-radio label="tech">{{ $t('tools.aiPicture.stTech') }}</el-radio>
            <el-radio label="cozy">{{ $t('tools.aiPicture.stCozy') }}</el-radio>
            <el-radio label="professional">{{ $t('tools.aiPicture.stPro') }}</el-radio>
            <el-radio label="creative">{{ $t('tools.aiPicture.stCreative') }}</el-radio>
            <el-radio label="vintage">{{ $t('tools.aiPicture.stVintage') }}</el-radio>
            <el-radio label="luxury">{{ $t('tools.aiPicture.stLuxury') }}</el-radio>
            <el-radio label="other">{{ $t('tools.aiPicture.other') }}</el-radio>
          </el-radio-group>
          <el-input v-if="stylePreference === 'other'" v-model="stylePreferenceCustom" :placeholder="$t('tools.aiPicture.phStyle')" class="w-full" />
        </div>
        
        <div class="mb-3">
          <div class="text-base font-medium mb-2">{{ $t('tools.aiPicture.secLayout') }}</div>
          <el-radio-group v-model="layoutDesign" class="mb-2">
            <el-radio label="single_column">{{ $t('tools.aiPicture.lySingle') }}</el-radio>
            <el-radio label="two_column">{{ $t('tools.aiPicture.lyDouble') }}</el-radio>
            <el-radio label="three_column">{{ $t('tools.aiPicture.lyTriple') }}</el-radio>
            <el-radio label="grid">{{ $t('tools.aiPicture.lyGrid') }}</el-radio>
            <el-radio label="gallery">{{ $t('tools.aiPicture.lyGallery') }}</el-radio>
            <el-radio label="masonry">{{ $t('tools.aiPicture.lyMasonry') }}</el-radio>
            <el-radio label="card">{{ $t('tools.aiPicture.lyCard') }}</el-radio>
            <el-radio label="sidebar">{{ $t('tools.aiPicture.lySidebar') }}</el-radio>
            <el-radio label="split_screen">{{ $t('tools.aiPicture.lySplit') }}</el-radio>
            <el-radio label="fullscreen_scroll">{{ $t('tools.aiPicture.lyFullScroll') }}</el-radio>
            <el-radio label="mixed">{{ $t('tools.aiPicture.lyMixed') }}</el-radio>
            <el-radio label="responsive">{{ $t('tools.aiPicture.lyResponsive') }}</el-radio>
            <el-radio label="other">{{ $t('tools.aiPicture.other') }}</el-radio>
          </el-radio-group>
          <el-input v-if="layoutDesign === 'other'" v-model="layoutDesignCustom" :placeholder="$t('tools.aiPicture.phLayout')" class="w-full" />
        </div>
        
        <div class="mb-3">
          <div class="text-base font-medium mb-2">{{ $t('tools.aiPicture.secColor') }}</div>
          <el-radio-group v-model="colorScheme" class="mb-2">
            <el-radio label="blue">{{ $t('tools.aiPicture.colBlue') }}</el-radio>
            <el-radio label="green">{{ $t('tools.aiPicture.colGreen') }}</el-radio>
            <el-radio label="red">{{ $t('tools.aiPicture.colRed') }}</el-radio>
            <el-radio label="grayscale">{{ $t('tools.aiPicture.colGray') }}</el-radio>
            <el-radio label="black_white">{{ $t('tools.aiPicture.colBw') }}</el-radio>
            <el-radio label="warm">{{ $t('tools.aiPicture.colWarm') }}</el-radio>
            <el-radio label="cool">{{ $t('tools.aiPicture.colCool') }}</el-radio>
            <el-radio label="other">{{ $t('tools.aiPicture.other') }}</el-radio>
          </el-radio-group>
          <el-input v-if="colorScheme === 'other'" v-model="colorSchemeCustom" :placeholder="$t('tools.aiPicture.phColor')" class="w-full" />
        </div>
      </div>
      
      <!-- 详细描述输入框 -->
      <div class="relative">
        <div class="text-base font-medium mb-2">{{ $t('tools.aiPicture.secDetail') }}</div>
        <el-input
            v-model="prompt"
            :maxlength="2000"
            :placeholder="placeholder"
            :rows="5"
            resize="none"
            type="textarea"
            @blur="handleBlur"
            @focus="handleFocus"
        />
        <div class="flex absolute right-2 bottom-2">
          <el-tooltip effect="light">
            <template #content>
              <div>
                {{ $t('tools.aiPicture.tooltipLicenseBefore') }}<a
                  class="text-blue-600"
                  href="https://plugin.LightningRAG.com/license"
                  target="_blank"
              >{{ $t('tools.aiPicture.buyLicense') }}</a>
              </div>
            </template>
            <el-button
                type="primary"
                @click="llmAutoFunc()"
            >
              <el-icon size="18">
                <ai-lrag/>
              </el-icon>
              {{ $t('tools.aiPicture.btnGenerate') }}
            </el-button>
          </el-tooltip>
        </div>
      </div>
    </div>
    <div>
      <div v-if="!outPut">
        <el-empty :image-size="200"/>
      </div>
      <div v-if="outPut && htmlFromLLM">
        <el-tabs type="border-card">
          <el-tab-pane :label="$t('tools.aiPicture.tabPreview')">
            <div class="h-[500px] overflow-auto bg-gray-50 p-4 rounded">
              <div v-if="!loadedComponents" class="text-gray-500 text-center py-4">
                {{ $t('tools.aiPicture.loadingComp') }}
              </div>
              <component
                v-else
                :is="loadedComponents" 
                class="vue-component-container w-full"
              />
            </div>
          </el-tab-pane>
          <el-tab-pane :label="$t('tools.aiPicture.tabSource')">
            <div class="relative h-[500px] overflow-auto bg-gray-50 p-4 rounded">
              <el-button 
                type="primary" 
                :icon="DocumentCopy" 
                class="absolute top-2 right-2 px-2 py-1" 
                @click="copySnippet(htmlFromLLM)" 
                plain
              >
                {{ $t('tools.aiPicture.btnCopy') }}
              </el-button>
              <pre class="mt-10 whitespace-pre-wrap">{{ htmlFromLLM }}</pre>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script setup>
import { llmAuto } from '@/api/autoCode'
import { ref, markRaw, computed, h } from 'vue'
import * as Vue from "vue";
import WarningBar from '@/components/warningBar/warningBar.vue'
import { ElMessage } from 'element-plus'
import { defineAsyncComponent } from 'vue'
import { DocumentCopy } from '@element-plus/icons-vue'
import { loadModule } from "vue3-sfc-loader";
import { useI18n } from 'vue-i18n'
import { i18n } from '@/locale/index.js'

defineOptions({
  name: 'Picture'
})

const { t } = useI18n()

const handleFocus = () => {
  document.addEventListener('keydown', handleKeydown);
}

const handleBlur = () => {
  document.removeEventListener('keydown', handleKeydown);
}

const handleKeydown = (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    llmAutoFunc()
  }
}

// 复制方法：把某个字符串写进剪贴板
const copySnippet = (vueString) => {
  navigator.clipboard.writeText(vueString)
      .then(() => {
        ElMessage({
          message: t('tools.aiPicture.copyOk'),
          type: 'success',
        })
      })
      .catch(err => {
        ElMessage({
          message: t('tools.aiPicture.copyFail'),
          type: 'warning',
        })
      })
}

const pageType = ref('corporate')
const pageTypeCustom = ref('')
const contentBlocks = ref(['banner', 'product_service'])
const contentBlocksCustom = ref('')
const stylePreference = ref('minimal')
const stylePreferenceCustom = ref('')
const layoutDesign = ref('responsive')
const layoutDesignCustom = ref('')
const colorScheme = ref('blue')
const colorSchemeCustom = ref('')

const pageTypeContentMap = {
  corporate: ['banner', 'product_service', 'features', 'cases', 'contact'],
  ecommerce: ['banner', 'product_list', 'product_card', 'cart', 'category', 'hot_picks', 'flash_sale', 'checkout', 'reviews'],
  blog: ['banner', 'news', 'reviews', 'contact'],
  product: ['banner', 'product_service', 'features', 'pricing', 'faq'],
  landing: ['banner', 'features', 'contact', 'stats']
}

const prompt = ref('')

// 判断是否返回的标志
const outPut = ref(false)
// 容纳llm返回的vue组件代码
const htmlFromLLM = ref("")

// 存储加载的组件
const loadedComponents = ref(null)

const loadVueComponent = async (vueCode) => {
  try {
    // 使用内存中的虚拟路径
    const fakePath = `virtual:component-0.vue`
    
    const component = defineAsyncComponent({
      loader: async () => {
        try {
          const options = {
            moduleCache: {
              vue: Vue,
            },
            getFile(url) {
              // 处理所有可能的URL格式，包括相对路径、绝对路径等
              // 提取路径的最后部分，忽略查询参数
              const fileName = url.split('/').pop().split('?')[0]
              const componentFileName = fakePath.split('/').pop()
              
              // 如果文件名包含我们的组件名称，或者url完全匹配fakePath
              if (fileName === componentFileName || url === fakePath || 
                  url === `./component/0.vue`) {
                return Promise.resolve({
                  type: '.vue',
                  getContentData: () => vueCode
                })
              }
              
              console.warn('Unknown file requested:', url)
              return Promise.reject(new Error(`File not found: ${url}`))
            },
            addStyle(textContent) {
              // 不再将样式添加到document.head，而是返回样式内容
              // 稍后会将样式添加到Shadow DOM中
              return textContent
            },
            handleModule(type, source, path, options) {
              // 默认处理器
              return undefined
            },
            log(type, ...args) {
              console.log(`[vue3-sfc-loader] [${type}]`, ...args)
            }
          }
          
          // 尝试加载组件
          const comp = await loadModule(fakePath, options)
          return comp.default || comp
        } catch (error) {
          console.error('Component load error:', error)
          throw error
        }
      },
      loadingComponent: {
        setup() {
          return () => h('div', i18n.global.t('tools.aiPicture.loaderInline'))
        }
      },
      errorComponent: {
        props: ['error'],
        setup(props) {
          console.error('Error component received:', props.error)
          return () =>
            h(
              'div',
              `${i18n.global.t('tools.aiPicture.errLoadComp')} ${props.error?.message || ''}`
            )
        }
      },
      // 添加超时和重试选项
      timeout: 30000,
      delay: 200,
      suspensible: false,
      onError(error, retry, fail) {
        console.error('Load error details:', error)
        fail()
      }
    })

    // 创建一个包装组件，使用Shadow DOM隔离样式
    const ShadowWrapper = {
      name: 'ShadowWrapper',
      setup() {
        return {}
      },
      render() {
        return Vue.h('div', { class: 'shadow-wrapper' })
      },
      mounted() {
        // 创建Shadow DOM
        const shadowRoot = this.$el.attachShadow({ mode: 'open' })
        
        // 创建一个容器元素
        const container = document.createElement('div')
        container.className = 'shadow-container'
        shadowRoot.appendChild(container)
        
        // 提取组件中的样式
        const styleContent = vueCode.match(/<style[^>]*>([\s\S]*?)<\/style>/i)?.[1] || ''
        
        // 创建样式元素并添加到Shadow DOM
        if (styleContent) {
          const style = document.createElement('style')
          style.textContent = styleContent
          shadowRoot.appendChild(style)
        }
        
        // 创建Vue应用并挂载到Shadow DOM容器中
        const app = Vue.createApp({
          render: () => Vue.h(component)
        })
        app.mount(container)
      }
    }

    loadedComponents.value = markRaw(ShadowWrapper)
    return ShadowWrapper
  } catch (error) {
    console.error('Component creation error:', error)
    return null
  }
}

// 当页面用途改变时，更新内容板块的选择
const handlePageTypeChange = (value) => {
  if (value !== 'other' && pageTypeContentMap[value]) {
    contentBlocks.value = [...pageTypeContentMap[value]]
  }
}

const resolveLabel = (key, i18nPrefix) => {
  const fullKey = `${i18nPrefix}.${key}`
  const translated = t(fullKey)
  return translated !== fullKey ? translated : key
}

const llmAutoFunc = async () => {
  let fullPrompt = ''

  const pageLabel = pageType.value === 'other'
    ? pageTypeCustom.value
    : resolveLabel(pageType.value, 'tools.aiPicture.val.pageType')
  fullPrompt += `${t('tools.aiPicture.promptPageUse')}: ${pageLabel}\n`

  fullPrompt += `${t('tools.aiPicture.promptBlocks')}: `
  const blocks = contentBlocks.value
    .filter((b) => b !== 'other')
    .map((b) => resolveLabel(b, 'tools.aiPicture.val.block'))
  if (contentBlocksCustom.value) {
    blocks.push(contentBlocksCustom.value)
  }
  fullPrompt += blocks.join(', ') + '\n'

  const styleLabel = stylePreference.value === 'other'
    ? stylePreferenceCustom.value
    : resolveLabel(stylePreference.value, 'tools.aiPicture.val.style')
  fullPrompt += `${t('tools.aiPicture.promptStyle')}: ${styleLabel}\n`

  const layoutLabel = layoutDesign.value === 'other'
    ? layoutDesignCustom.value
    : resolveLabel(layoutDesign.value, 'tools.aiPicture.val.layout')
  fullPrompt += `${t('tools.aiPicture.promptLayout')}: ${layoutLabel}\n`

  const colorLabel = colorScheme.value === 'other'
    ? colorSchemeCustom.value
    : resolveLabel(colorScheme.value, 'tools.aiPicture.val.color')
  fullPrompt += `${t('tools.aiPicture.promptColor')}: ${colorLabel}\n`

  if (prompt.value) {
    fullPrompt += `\n${t('tools.aiPicture.promptDetail')}: ${prompt.value}`
  }
  
  const res = await llmAuto({web: fullPrompt, mode: 'createWeb'})
  if (res.code === 0) {
    outPut.value = true
    // 添加返回的Vue组件代码到数组
    htmlFromLLM.value = res.data.text
    // 加载新生成的组件
    await loadVueComponent(res.data.text)
  }
}

const placeholder = computed(() => t('tools.aiPicture.textareaPh'))
</script>
