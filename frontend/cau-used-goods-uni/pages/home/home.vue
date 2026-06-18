<template>
  <view class="page">
    <view class="hero">
      <text class="eyebrow">CAU CAMPUS MARKET</text>
      <text class="headline">让闲置，在校园里重新发光</text>
      <view class="search" @click="goSearch">搜索教材、数码、生活用品</view>
    </view>

    <view class="section-head">
      <text class="section-title">逛分类</text>
      <text class="muted">点一级分类查看商品</text>
    </view>

    <scroll-view scroll-x class="category-scroll">
      <view class="category-row">
        <view
          v-for="item in primaryCategories"
          :key="item.id"
          class="category-pill"
          :class="{ active: Number(item.id) === Number(activePrimaryId) }"
          @click="selectPrimary(item.id)"
        >
          {{ item.name }}
        </view>
      </view>
    </scroll-view>

    <scroll-view v-if="activePrimaryId && childCategories.length" scroll-x class="sub-category-scroll">
      <view class="sub-category-row">
        <view
          v-for="item in childCategories"
          :key="item.id"
          class="sub-category"
          @click="goSearchCategory(item)"
        >
          {{ item.name }}
        </view>
      </view>
    </scroll-view>

    <view class="section-head">
      <text class="section-title">{{ activePrimaryName || '新鲜发布' }}</text>
      <text class="muted">下拉刷新最新上架</text>
    </view>

    <view v-if="products.length" class="grid">
      <ProductCard v-for="item in products" :key="item.id" :product="item" />
    </view>

    <view v-else-if="!loading" class="empty-state">暂时没有在售商品</view>
    <view class="load-state">{{ loading ? '正在加载...' : finished ? '已经到底啦' : '' }}</view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import ProductCard from '../../components/ProductCard.vue'
import { listCategories, listProducts } from '../../api/product'
import { buildCategoryMap, formatProduct } from '../../utils/product-format'

const categories = ref([])
const rawProducts = ref([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const activePrimaryId = ref(0)

const fallbackChildren = {
  教材资料: ['课本', '考研资料', '四六级', '教辅资料', '笔记'],
  电子产品: ['手机', '耳机', '键盘', '平板', '电脑', '充电器'],
  生活用品: ['收纳', '台灯', '床品', '餐具', '清洁用品'],
  服饰鞋包: ['上衣', '裤子', '鞋子', '包包', '配饰'],
  运动户外: ['篮球', '羽毛球', '健身', '骑行', '户外装备'],
  其他: ['其他']
}

const parentIdOf = (item) => Number(item.parentId || item.parent_id || 0)
const primaryCategories = computed(() => categories.value.filter((item) => Number(item.id) !== 0 && parentIdOf(item) === 0))
const childCategories = computed(() => {
  if (!activePrimaryId.value) return []
  const realChildren = categories.value.filter((item) => {
    return parentIdOf(item) === Number(activePrimaryId.value) && Number(item.id) !== Number(activePrimaryId.value)
  })
  if (realChildren.length) return realChildren

  const primary = primaryCategories.value.find((item) => Number(item.id) === Number(activePrimaryId.value))
  return (fallbackChildren[primary?.name] || []).map((name) => ({
    id: `keyword-${name}`,
    name,
    keywordOnly: true
  }))
})
const categoryMap = computed(() => buildCategoryMap(categories.value))
const products = computed(() => rawProducts.value.map((item) => formatProduct(item, categoryMap.value)))
const finished = computed(() => rawProducts.value.length >= total.value && total.value > 0)
const activePrimaryName = computed(() => primaryCategories.value.find((item) => Number(item.id) === Number(activePrimaryId.value))?.name || '')
const activeCategoryIds = computed(() => {
  if (!activePrimaryId.value) return []
  const realChildIds = childCategories.value
    .filter((item) => !item.keywordOnly)
    .map((item) => item.id)
  return [activePrimaryId.value, ...realChildIds]
})

const loadCategories = async () => {
  categories.value = await listCategories()
}

const loadProducts = async (reset = false) => {
  if (loading.value || (!reset && finished.value)) return

  loading.value = true
  try {
    const nextPage = reset ? 1 : page.value
    const baseParams = {
      page: nextPage,
      pageSize: 8,
      sort: 'newest'
    }
    const requests = activeCategoryIds.value.length
      ? activeCategoryIds.value.map((categoryId) => listProducts({ ...baseParams, categoryId }))
      : [listProducts(baseParams)]
    const results = await Promise.all(requests)

    const merged = []
    const seen = {}
    results.forEach((result) => {
      ;(result.list || []).forEach((item) => {
        if (seen[item.id]) return
        seen[item.id] = true
        merged.push(item)
      })
    })
    merged.sort((a, b) => String(b.createTime || '').localeCompare(String(a.createTime || '')))
    rawProducts.value = reset ? merged : rawProducts.value.concat(merged)
    total.value = results.reduce((sum, result) => sum + Number(result.total || 0), 0)
    page.value = nextPage + 1
  } catch (error) {
    uni.showToast({ title: error.message || '商品加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const refresh = async () => {
  try {
    await loadCategories()
    await loadProducts(true)
  } finally {
    uni.stopPullDownRefresh()
  }
}

const goSearch = () => uni.navigateTo({ url: '/pages/search/search' })
const selectPrimary = (categoryId) => {
  activePrimaryId.value = Number(activePrimaryId.value) === Number(categoryId) ? 0 : Number(categoryId)
  loadProducts(true)
}
const goSearchCategory = (category) => {
  if (category.keywordOnly) {
    uni.navigateTo({ url: `/pages/search/search?keyword=${encodeURIComponent(category.name)}` })
    return
  }
  uni.navigateTo({ url: `/pages/search/search?categoryId=${category.id}&keyword=${encodeURIComponent(category.name)}` })
}

onShow(() => {
  if (!rawProducts.value.length || uni.getStorageSync('PRODUCT_LIST_DIRTY')) {
    uni.removeStorageSync('PRODUCT_LIST_DIRTY')
    refresh()
  }
})

onReachBottom(() => loadProducts())
onPullDownRefresh(refresh)
</script>

<style scoped>
.page { min-height: 100vh; padding-bottom: 36rpx; }
.hero { padding: 92rpx 30rpx 34rpx; border-radius: 0 0 40rpx 40rpx; background: linear-gradient(145deg, #1f6a49, #328660); color: #fff; }
.eyebrow, .headline { display: block; }
.eyebrow { color: rgba(255,255,255,.7); font-size: 20rpx; letter-spacing: 3rpx; }
.headline { margin-top: 14rpx; font-size: 38rpx; font-weight: 700; }
.search { margin-top: 30rpx; padding: 24rpx; border-radius: 20rpx; background: #fff; color: #9ca7a3; }
.section-head { display: flex; align-items: center; justify-content: space-between; padding: 32rpx 28rpx 18rpx; }
.section-title { color: #26342f; font-size: 32rpx; font-weight: 700; }
.muted, .load-state { color: #929c98; font-size: 23rpx; }
.category-scroll, .sub-category-scroll { width: 100%; white-space: nowrap; }
.category-row, .sub-category-row { display: inline-flex; }
.category-row { gap: 16rpx; padding: 2rpx 28rpx 8rpx; }
.category-pill { padding: 16rpx 28rpx; border-radius: 999rpx; background: #fff; color: #23734f; font-size: 26rpx; box-shadow: 0 4rpx 14rpx rgba(28,68,52,.06); }
.category-pill.active { background: #23734f; color: #fff; font-weight: 700; }
.sub-category-row { gap: 14rpx; padding: 12rpx 28rpx 6rpx; }
.sub-category { padding: 12rpx 22rpx; border-radius: 999rpx; background: #edf4f1; color: #23734f; font-size: 24rpx; }
.grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20rpx; padding: 0 28rpx; }
.empty-state { margin: 80rpx 28rpx 0; color: #929c98; font-size: 28rpx; text-align: center; }
.load-state { padding: 28rpx; text-align: center; }
</style>
