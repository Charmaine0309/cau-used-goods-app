<template>
  <view class="page">
    <view class="section-title">商品上下架</view>
    <view v-if="products.length === 0" class="empty">暂无商品</view>
    <view v-for="item in products" :key="item.id" class="card" @click="goStatus(item.id)">
      <view class="card-main">
        <view class="name">{{ item.title }}</view>
        <view class="desc">￥{{ item.price }} · {{ statusText(item.status) }}</view>
      </view>
      <view class="arrow">›</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAdminProducts } from '../../api/admin'

const products = ref([])

const load = async () => {
  try {
    const result = await getAdminProducts()
    products.value = result?.list || result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  }
}

onShow(load)

const statusText = (status) => {
  const map = { ON_SALE: '在售', OFF_SHELF: '已下架', LOCKED: '交易锁定', SOLD: '已售出', DELETED: '已删除' }
  return map[status] || status || '未知'
}

const goStatus = (id) => {
  uni.navigateTo({
    url: `/pages/admin-product-status/admin-product-status?id=${id}`
  })
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.card { display: flex; align-items: center; }
.card-main { flex: 1; min-width: 0; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.desc, .empty { margin-top: 8rpx; color: #667085; font-size: 26rpx; }
.arrow { margin-left: 20rpx; color: #b2bdca; font-size: 42rpx; }
</style>
