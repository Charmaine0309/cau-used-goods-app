<template>
  <view class="page">
    <view class="section-title">数据统计</view>

    <view class="grid">
      <view class="stat-card">
        <view class="num">{{ userOverview.totalUsers || 0 }}</view>
        <view class="label">用户总数</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ userOverview.verifiedUsers || 0 }}</view>
        <view class="label">已认证用户</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ productOverview.totalProducts || 0 }}</view>
        <view class="label">商品总数</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ productOverview.onSaleProducts || 0 }}</view>
        <view class="label">在售商品</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ orderOverview.totalOrders || 0 }}</view>
        <view class="label">订单总数</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ reportOverview.totalReports || 0 }}</view>
        <view class="label">举报总数</view>
      </view>
    </view>

    <view class="section-title">商品分类分布</view>
    <view v-if="categoryList.length === 0" class="empty">暂无分类统计</view>
    <view v-for="item in categoryList" :key="item.categoryId || item.categoryName" class="row-card">
      <text>{{ item.categoryName || item.name || '分类' }}</text>
      <text>{{ item.count || 0 }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  getCategoryDistribution,
  getOrderOverview,
  getProductOverview,
  getReportOverview,
  getUserOverview
} from '../../api/admin'

const userOverview = ref({})
const productOverview = ref({})
const orderOverview = ref({})
const reportOverview = ref({})
const categoryList = ref([])

const load = async () => {
  try {
    const [users, products, orders, reports, categories] = await Promise.all([
      getUserOverview(),
      getProductOverview(),
      getOrderOverview(),
      getReportOverview(),
      getCategoryDistribution()
    ])
    userOverview.value = users || {}
    productOverview.value = products || {}
    orderOverview.value = orders || {}
    reportOverview.value = reports || {}
    categoryList.value = categories?.list || categories || []
  } catch (error) {
    uni.showToast({ title: error.message || '统计加载失败', icon: 'none' })
  }
}

onShow(load)
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 18rpx; }
.stat-card, .row-card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; }
.num { font-size: 42rpx; font-weight: 700; color: #17a84b; }
.label, .empty { margin-top: 8rpx; color: #667085; font-size: 26rpx; }
.row-card { display: flex; justify-content: space-between; margin-bottom: 14rpx; color: #475467; font-size: 28rpx; }
</style>
