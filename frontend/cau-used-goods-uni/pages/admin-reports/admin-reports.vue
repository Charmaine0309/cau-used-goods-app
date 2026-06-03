<template>
  <view class="page">
    <view class="page-title">举报处理</view>

    <view class="summary">
      <view class="summary-main">
        <view class="summary-number">{{ pendingCount }}</view>
        <view class="summary-label">待处理举报</view>
      </view>
      <view :class="['summary-status', pendingCount ? 'warning' : 'safe']">
        {{ pendingCount ? '需要处理' : '暂无风险' }}
      </view>
    </view>

    <view class="filter-row">
      <view
        v-for="item in targetFilters"
        :key="item.value"
        :class="['filter-chip', activeTarget === item.value ? 'active' : '']"
        @click="activeTarget = item.value"
      >
        {{ item.label }} {{ countByTarget(item.value) }}
      </view>
    </view>

    <view class="type-board">
      <view class="type-item">
        <text>商品举报</text>
        <text>{{ countByTarget('PRODUCT') }}</text>
      </view>
      <view class="type-item">
        <text>用户举报</text>
        <text>{{ countByTarget('USER') }}</text>
      </view>
      <view class="type-item">
        <text>订单举报</text>
        <text>{{ countByTarget('ORDER') }}</text>
      </view>
    </view>

    <view v-if="filteredReports.length === 0" class="empty">暂无对应举报</view>
    <view v-for="item in filteredReports" :key="item.id" class="report-card">
      <view class="card-head">
        <view>
          <view class="report-title">{{ reasonText(item.reasonType) }}</view>
          <view class="report-sub">{{ targetText(item.targetType) }} #{{ item.targetId }}</view>
        </view>
        <view :class="['status-badge', item.status]">{{ statusText(item.status) }}</view>
      </view>

      <view class="report-desc">{{ item.description || '暂无补充说明' }}</view>

      <view class="meta-row">
        <text>举报人：{{ item.reporterNickname || `用户${item.reporterId}` }}</text>
        <text>{{ shortTime(item.createTime) }}</text>
      </view>

      <view v-if="canHandle(item.status)" class="actions">
        <button size="mini" class="pass" @click="handleReport(item.id, 'RESOLVED')">处理完成</button>
        <button size="mini" class="reject" @click="handleReport(item.id, 'REJECTED')">驳回</button>
      </view>
      <view v-else-if="item.handleResult" class="handle-result">{{ item.handleResult }}</view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAdminReports, handleAdminReport } from '../../api/admin'

const reports = ref([])
const activeTarget = ref('ALL')

const targetFilters = [
  { label: '全部', value: 'ALL' },
  { label: '商品', value: 'PRODUCT' },
  { label: '用户', value: 'USER' },
  { label: '订单', value: 'ORDER' }
]

const load = async () => {
  try {
    const result = await getAdminReports()
    reports.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  }
}

onShow(load)

const pendingCount = computed(() => {
  return reports.value.filter((item) => ['PENDING', 'PROCESSING'].includes(item.status)).length
})

const filteredReports = computed(() => {
  if (activeTarget.value === 'ALL') return reports.value
  return reports.value.filter((item) => item.targetType === activeTarget.value)
})

const countByTarget = (targetType) => {
  if (targetType === 'ALL') return reports.value.length
  return reports.value.filter((item) => item.targetType === targetType).length
}

const targetText = (targetType) => {
  const map = {
    PRODUCT: '商品',
    USER: '用户',
    ORDER: '订单'
  }
  return map[targetType] || targetType || '对象'
}

const reasonText = (reasonType) => {
  const map = {
    FAKE: '虚假信息',
    FRAUD: '疑似诈骗',
    PROHIBITED: '违规商品',
    INAPPROPRIATE: '不当内容',
    HARASSMENT: '骚扰行为',
    OTHER: '其他原因'
  }
  return map[reasonType] || reasonType || '举报'
}

const statusText = (status) => {
  const map = {
    PENDING: '待处理',
    PROCESSING: '处理中',
    RESOLVED: '已处理',
    REJECTED: '已驳回',
    CLOSED: '已关闭'
  }
  return map[status] || status || '未知'
}

const shortTime = (value) => {
  if (!value) return ''
  return String(value).replace('T', ' ').slice(0, 16)
}

const canHandle = (status) => {
  return ['PENDING', 'PROCESSING'].includes(status)
}

const handleReport = async (id, status) => {
  try {
    await handleAdminReport(id, status)
    uni.showToast({ title: '举报已处理', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '举报处理失败', icon: 'none' })
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 28rpx 24rpx;
  background: #f5f6f8;
  box-sizing: border-box;
}

.page-title {
  margin: 18rpx 0 24rpx;
  font-size: 38rpx;
  font-weight: 700;
  color: #1f2933;
}

.summary {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 30rpx;
  border-radius: 18rpx;
  background: #fff;
}

.summary-number {
  font-size: 54rpx;
  line-height: 60rpx;
  font-weight: 700;
  color: #ef4444;
}

.summary-label {
  margin-top: 10rpx;
  font-size: 26rpx;
  color: #667085;
}

.summary-status {
  padding: 8rpx 16rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
}

.summary-status.warning {
  background: #fee2e2;
  color: #ef4444;
}

.summary-status.safe {
  background: #dcfce7;
  color: #16a34a;
}

.filter-row {
  display: flex;
  gap: 14rpx;
  margin: 24rpx 0;
  overflow-x: auto;
}

.filter-chip {
  flex-shrink: 0;
  padding: 14rpx 22rpx;
  border-radius: 999rpx;
  background: #fff;
  color: #667085;
  font-size: 26rpx;
}

.filter-chip.active {
  background: #17a84b;
  color: #fff;
  font-weight: 700;
}

.type-board {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14rpx;
  margin-bottom: 24rpx;
}

.type-item {
  padding: 22rpx 16rpx;
  border-radius: 14rpx;
  background: #fff;
  text-align: center;
  font-size: 24rpx;
  color: #667085;
}

.type-item text:last-child {
  display: block;
  margin-top: 10rpx;
  font-size: 34rpx;
  font-weight: 700;
  color: #1f2933;
}

.report-card,
.empty {
  padding: 28rpx;
  border-radius: 16rpx;
  background: #fff;
  margin-bottom: 18rpx;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.report-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2933;
}

.report-sub {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.status-badge {
  flex-shrink: 0;
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #eef2f6;
  color: #667085;
  font-size: 22rpx;
}

.status-badge.PENDING,
.status-badge.PROCESSING {
  background: #fee2e2;
  color: #ef4444;
}

.status-badge.RESOLVED {
  background: #dcfce7;
  color: #16a34a;
}

.report-desc,
.empty {
  margin-top: 18rpx;
  color: #667085;
  font-size: 26rpx;
  line-height: 38rpx;
}

.meta-row {
  display: flex;
  justify-content: space-between;
  gap: 20rpx;
  margin-top: 20rpx;
  font-size: 22rpx;
  color: #98a2b3;
}

.actions {
  margin-top: 22rpx;
  display: flex;
  gap: 18rpx;
}

.pass {
  background: #17a84b;
  color: #fff;
}

.reject {
  background: #fff1f2;
  color: #ef4444;
}

.handle-result {
  margin-top: 18rpx;
  padding: 16rpx;
  border-radius: 12rpx;
  background: #f8fafc;
  color: #667085;
  font-size: 24rpx;
}
</style>
