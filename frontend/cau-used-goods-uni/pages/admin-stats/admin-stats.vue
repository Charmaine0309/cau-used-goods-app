<template>
  <view class="page">
    <view class="section-title">申诉统计</view>
    <view class="summary-card">
      <view class="summary-main">
        <view>
          <view class="summary-num">{{ appealOverview.totalAppeals || 0 }}</view>
          <view class="summary-label">申诉总数</view>
        </view>
        <view class="summary-badge" :class="{ danger: hasPendingAppeals }">
          {{ hasPendingAppeals ? '需要处理' : '暂无待办' }}
        </view>
      </view>

      <view class="progress-row">
        <view class="progress-head">
          <text>已通过</text>
          <text>{{ percent(appealOverview.approvedAppeals, appealOverview.totalAppeals) }}</text>
        </view>
        <view class="progress-track">
          <view class="progress-bar success" :style="barStyle(appealOverview.approvedAppeals, appealOverview.totalAppeals)"></view>
        </view>
      </view>

      <view class="progress-row">
        <view class="progress-head">
          <text>已驳回</text>
          <text>{{ percent(appealOverview.rejectedAppeals, appealOverview.totalAppeals) }}</text>
        </view>
        <view class="progress-track">
          <view class="progress-bar reject" :style="barStyle(appealOverview.rejectedAppeals, appealOverview.totalAppeals)"></view>
        </view>
      </view>
    </view>

    <view class="section-title">状态分布</view>
    <view class="list-card">
      <view v-for="item in appealStatusRows" :key="item.label" class="row-card">
        <view class="row-left">
          <text class="dot" :class="item.type"></text>
          <text>{{ item.label }}</text>
        </view>
        <text>{{ item.value }}</text>
      </view>
    </view>

    <view class="section-title">申诉对象</view>
    <view class="list-card">
      <view v-for="item in appealTargetRows" :key="item.label" class="row-card">
        <text>{{ item.label }}</text>
        <text>{{ item.value }}</text>
      </view>
    </view>

    <view class="section-title">商品分类分布</view>
    <view v-if="categoryList.length === 0" class="empty">暂无分类统计</view>
    <view v-for="item in categoryList" :key="item.categoryId || item.categoryName" class="row-card standalone">
      <text>{{ item.categoryName || item.name || '分类' }}</text>
      <text>{{ item.productCount || item.count || 0 }}</text>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  getAppealOverview,
  getCategoryDistribution
} from '../../api/admin'

const appealOverview = ref({})
const categoryList = ref([])

const hasPendingAppeals = computed(() => Number(appealOverview.value.pendingAppeals || 0) > 0)

const appealStatusRows = computed(() => [
  { label: '待处理', value: appealOverview.value.pendingAppeals || 0, type: 'warning' },
  { label: '处理中', value: appealOverview.value.processingAppeals || 0, type: 'info' },
  { label: '已通过', value: appealOverview.value.approvedAppeals || 0, type: 'success' },
  { label: '已驳回', value: appealOverview.value.rejectedAppeals || 0, type: 'reject' },
  { label: '已关闭', value: appealOverview.value.closedAppeals || 0, type: 'muted' }
])

const appealTargetRows = computed(() => [
  { label: '商品申诉', value: appealOverview.value.productAppeals || 0 },
  { label: '用户申诉', value: appealOverview.value.userAppeals || 0 },
  { label: '订单申诉', value: appealOverview.value.orderAppeals || 0 },
  { label: '举报申诉', value: appealOverview.value.reportAppeals || 0 }
])

const load = async () => {
  try {
    const [appeals, categories] = await Promise.all([
      getAppealOverview(),
      getCategoryDistribution()
    ])
    appealOverview.value = appeals || {}
    categoryList.value = categories?.list || categories || []
  } catch (error) {
    uni.showToast({ title: error.message || '统计加载失败', icon: 'none' })
  }
}

onShow(load)

const percent = (value, total) => {
  const num = Number(value || 0)
  const den = Number(total || 0)
  if (!den) return '0%'
  return `${Math.round((num / den) * 100)}%`
}

const barStyle = (value, total) => {
  const num = Number(value || 0)
  const den = Number(total || 0)
  if (!den) return 'width: 0%;'
  return `width: ${Math.round((num / den) * 100)}%;`
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f5f6f8;
  box-sizing: border-box;
}

.section-title {
  margin: 24rpx 0 18rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2933;
}

.summary-card,
.list-card,
.empty,
.row-card.standalone {
  padding: 28rpx;
  border-radius: 16rpx;
  background: #fff;
  box-sizing: border-box;
}

.empty {
  margin-top: 8rpx;
  color: #667085;
  font-size: 26rpx;
}

.summary-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
  margin-bottom: 28rpx;
}

.summary-num {
  font-size: 64rpx;
  line-height: 70rpx;
  font-weight: 700;
  color: #1f2933;
}

.summary-label {
  margin-top: 8rpx;
  font-size: 26rpx;
  color: #667085;
}

.summary-badge {
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  background: #dcfce7;
  color: #16a34a;
  font-size: 24rpx;
}

.summary-badge.danger {
  background: #fee2e2;
  color: #ef4444;
}

.progress-row {
  margin-top: 20rpx;
}

.progress-head,
.row-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  color: #475467;
  font-size: 28rpx;
}

.progress-head {
  margin-bottom: 12rpx;
}

.progress-track {
  height: 14rpx;
  border-radius: 999rpx;
  background: #eef2f6;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  border-radius: 999rpx;
}

.progress-bar.success {
  background: #17a84b;
}

.progress-bar.reject {
  background: #ef4444;
}

.list-card {
  padding: 0;
  overflow: hidden;
}

.row-card {
  min-height: 96rpx;
  padding: 24rpx 28rpx;
  border-bottom: 1rpx solid #eef0f3;
  box-sizing: border-box;
}

.row-card:last-child {
  border-bottom: 0;
}

.row-card.standalone {
  margin-bottom: 14rpx;
}

.row-left {
  display: flex;
  align-items: center;
  gap: 14rpx;
}

.dot {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  background: #98a2b3;
  flex-shrink: 0;
}

.dot.warning {
  background: #f59e0b;
}

.dot.info {
  background: #3b82f6;
}

.dot.success {
  background: #17a84b;
}

.dot.reject {
  background: #ef4444;
}

.dot.muted {
  background: #98a2b3;
}
</style>
