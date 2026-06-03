<template>
  <view class="page">
    <view class="top-row">
      <view class="page-title">CAU二手交易平台</view>
      <button class="mine-button" size="mini" @click="goMine">我的</button>
    </view>

    <view class="search-box" @click="goSearch">
      搜索二手商品
    </view>

    <view class="section-title">商品分类</view>

    <view class="category-list">
      <view class="category-item" @click="goCategory(1)">教材资料</view>
      <view class="category-item" @click="goCategory(2)">电子产品</view>
      <view class="category-item" @click="goCategory(3)">生活用品</view>
      <view class="category-item" @click="goCategory(5)">运动户外</view>
    </view>

    <view class="section-title">最新商品</view>

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

const goodsList = ref([])
const loading = ref(false)

const categoryNameMap = {
  1: '教材资料',
  2: '电子产品',
  3: '生活用品',
  4: '服饰鞋包',
  5: '运动户外',
  6: '其他'
}

const loadProducts = async () => {
  loading.value = true
  try {
    const result = await listProducts({
      status: 'ON_SALE',
      page: 1,
      pageSize: 20,
      sort: 'newest'
    })
    goodsList.value = (result?.list || []).map((item) => ({
      ...item,
      image: item.images?.[0] || '',
      categoryName: categoryNameMap[item.categoryId] || '商品'
    }))
  } catch (error) {
    uni.showToast({ title: error.message || '商品加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

onShow(loadProducts)

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

const goSearch = () => {
  uni.navigateTo({
    url: '/pages/search/search'
  })
}

const goCategory = (categoryId) => {
  uni.navigateTo({
    url: `/pages/category/category?categoryId=${categoryId}`
  })
}

const goDetail = (id) => {
  uni.navigateTo({
    url: `/pages/detail/detail?id=${id}`
  })
}

const goMine = () => {
  uni.navigateTo({
    url: '/pages/index/index'
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

.top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.page-title {
  font-size: 38rpx;
  font-weight: 700;
  color: #1f2933;
}

.mine-button {
  background: #ffffff;
  color: #17a84b;
}

.search-box {
  height: 72rpx;
  line-height: 72rpx;
  padding: 0 28rpx;
  border-radius: 36rpx;
  background: #ffffff;
  color: #9ca3af;
  font-size: 28rpx;
}

.section-title {
  margin-top: 36rpx;
  margin-bottom: 20rpx;
  font-size: 34rpx;
  font-weight: 700;
  color: #1f2933;
}

.category-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
}

.category-item {
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 16rpx;
  background: #ffffff;
  text-align: center;
  font-size: 28rpx;
  color: #374151;
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
