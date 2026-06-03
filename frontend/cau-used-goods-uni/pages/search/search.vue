<template>
  <view class="page">
    <view class="search-row">
      <input class="search-input" v-model="keyword" placeholder="请输入关键词" confirm-type="search" @confirm="loadProducts" />
      <button class="search-button" @click="loadProducts">搜索</button>
    </view>

    <view class="filter-row">
      <view
        v-for="item in sortOptions"
        :key="item.value"
        :class="['filter-item', sort === item.value ? 'active' : '']"
        @click="changeSort(item.value)"
      >
        {{ item.label }}
      </view>
    </view>

    <view class="goods-list">
      <view
        class="goods-card"
        v-for="item in goodsList"
        :key="item.id"
        @click="goDetail(item.id)"
      >
        <image v-if="item.image" class="goods-image" :src="item.image" mode="aspectFill" />
        <view v-else class="goods-image placeholder">{{ item.categoryName || '商品' }}</view>
        <view class="goods-info">
          <view class="goods-title">{{ item.title }}</view>
          <view class="goods-desc">{{ conditionText(item.conditionLevel) }}</view>
          <view class="goods-price">￥{{ item.price }}</view>
        </view>
      </view>
    </view>

    <view v-if="!loading && goodsList.length === 0" class="empty">暂无在售商品</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { listProducts } from '../../api/product'

const keyword = ref('')
const goodsList = ref([])
const loading = ref(false)
const sort = ref('newest')

const categoryNameMap = {
  1: '教材资料',
  2: '电子产品',
  3: '生活用品',
  4: '服饰鞋包',
  5: '运动户外',
  6: '其他'
}

const sortOptions = [
  { label: '最新', value: 'newest' },
  { label: '低价', value: 'price_asc' },
  { label: '高价', value: 'price_desc' }
]

onShow(loadProducts)

const loadProducts = async () => {
  loading.value = true
  try {
    const result = await listProducts({
      keyword: keyword.value,
      status: 'ON_SALE',
      page: 1,
      pageSize: 20,
      sort: sort.value
    })
    goodsList.value = (result?.list || []).map((item) => ({
      ...item,
      image: item.images?.[0] || '',
      categoryName: categoryNameMap[item.categoryId] || '商品'
    }))
  } catch (error) {
    uni.showToast({ title: error.message || '搜索失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const changeSort = (value) => {
  if (sort.value === value) return
  sort.value = value
  loadProducts()
}

const conditionText = (level) => {
  const map = {
    NEW: '全新',
    LIKE_NEW: '九成新',
    GOOD: '八成新',
    FAIR: '七成新',
    OLD: '旧物'
  }
  return map[level] || level || '成色未填写'
}

const goDetail = (id) => {
  uni.navigateTo({
    url: `/pages/detail/detail?id=${id}`
  })
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 32rpx;
  background: #f6f7f9;
  box-sizing: border-box;
}

.search-row {
  display: flex;
  gap: 16rpx;
}

.search-input {
  flex: 1;
  height: 76rpx;
  padding: 0 24rpx;
  border-radius: 12rpx;
  background: #ffffff;
  font-size: 28rpx;
}

.search-button {
  width: 140rpx;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: 12rpx;
  background: #1aad19;
  color: #ffffff;
  font-size: 28rpx;
}

.filter-row {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
  margin-bottom: 24rpx;
}

.filter-item {
  flex: 1;
  height: 64rpx;
  line-height: 64rpx;
  border-radius: 12rpx;
  background: #ffffff;
  text-align: center;
  font-size: 26rpx;
  color: #374151;
}

.filter-item.active {
  background: #1aad19;
  color: #ffffff;
}

.goods-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.goods-card {
  display: flex;
  padding: 20rpx;
  border-radius: 16rpx;
  background: #ffffff;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 12rpx;
  background: #d1d5db;
  flex-shrink: 0;
}

.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #7b8794;
  font-size: 24rpx;
}

.goods-info {
  margin-left: 24rpx;
  flex: 1;
}

.goods-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2933;
}

.goods-desc {
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #6b7280;
}

.goods-price {
  margin-top: 20rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: #e11d48;
}

.empty {
  margin-top: 40rpx;
  text-align: center;
  color: #98a2b3;
  font-size: 28rpx;
}
</style>
