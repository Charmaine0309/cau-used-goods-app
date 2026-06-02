<template>
  <view class="page">
    <view class="section-title">数据看板</view>
    <view class="grid">
      <view class="stat-card">
        <view class="num">{{ userOverview.totalUsers || 0 }}</view>
        <view class="label">用户总数</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ userOverview.pendingUsers || 0 }}</view>
        <view class="label">待认证</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ productOverview.totalProducts || 0 }}</view>
        <view class="label">商品总数</view>
      </view>
      <view class="stat-card">
        <view class="num">{{ orderOverview.totalOrders || 0 }}</view>
        <view class="label">订单总数</view>
      </view>
    </view>

    <view class="section-title">学生认证审核</view>
    <view v-if="verifications.length === 0" class="empty">暂无待审核认证</view>
    <view v-for="item in verifications" :key="item.id" class="card">
      <view class="name">{{ item.realName }} {{ item.studentId }}</view>
      <view class="desc">{{ item.college }}</view>
      <view class="actions">
        <button size="mini" class="pass" @click="review(item.id, 'VERIFIED')">通过</button>
        <button size="mini" class="reject" @click="review(item.id, 'REJECTED')">驳回</button>
      </view>
    </view>

    <view class="section-title">管理员日志</view>
    <view v-if="logs.length === 0" class="empty">暂无日志</view>
    <view v-for="log in logs" :key="log.id" class="log-item">
      {{ log.operationType || '操作' }} - {{ log.targetType || '对象' }}
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  getAdminLogs,
  getOrderOverview,
  getPendingStudentVerifications,
  getProductOverview,
  getUserOverview,
  reviewStudentVerification
} from '../../api/admin'

const userOverview = ref({})
const productOverview = ref({})
const orderOverview = ref({})
const verifications = ref([])
const logs = ref([])

const load = async () => {
  try {
    const [users, products, orders, pending, logResult] = await Promise.all([
      getUserOverview(),
      getProductOverview(),
      getOrderOverview(),
      getPendingStudentVerifications(),
      getAdminLogs()
    ])
    userOverview.value = users || {}
    productOverview.value = products || {}
    orderOverview.value = orders || {}
    verifications.value = pending?.items || []
    logs.value = logResult?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '后台数据加载失败', icon: 'none' })
  }
}

onShow(load)

const review = async (id, authStatus) => {
  try {
    await reviewStudentVerification(id, {
      authStatus,
      description: authStatus === 'VERIFIED' ? '学生认证审核通过' : '认证信息不符合要求'
    })
    uni.showToast({ title: '审核完成', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '审核失败', icon: 'none' })
  }
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 18rpx; }
.stat-card, .card, .log-item, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; }
.num { font-size: 42rpx; font-weight: 700; color: #17a84b; }
.label, .desc, .empty { margin-top: 8rpx; color: #667085; font-size: 26rpx; }
.card { margin-bottom: 18rpx; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.actions { margin-top: 18rpx; display: flex; gap: 18rpx; }
.pass { background: #17a84b; color: #fff; }
.reject { background: #fff1f2; color: #ef4444; }
.log-item { margin-bottom: 14rpx; color: #475467; font-size: 26rpx; }
</style>
