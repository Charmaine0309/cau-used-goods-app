<template>
  <view class="page">
    <view class="section-title">用户管理</view>
    <view v-if="users.length === 0" class="empty">暂无用户</view>
    <view v-for="item in users" :key="item.id" class="card">
      <view class="name">{{ item.nickname || item.realName || '微信用户' }}</view>
      <view class="desc">用户ID：{{ item.id }} · {{ roleText(item.role) }} · {{ authText(item.authStatus) }}</view>
      <view class="desc">手机号：{{ item.phone || '未填写' }}</view>
      <view class="desc">学院：{{ item.college || '未认证' }}</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAdminUsers } from '../../api/admin'

const users = ref([])

onShow(async () => {
  try {
    const result = await getAdminUsers()
    users.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '用户加载失败', icon: 'none' })
  }
})

const roleText = (role) => role === 'ADMIN' ? '管理员' : '普通用户'
const authText = (status) => {
  const map = { UNVERIFIED: '未认证', PENDING: '审核中', VERIFIED: '已认证', REJECTED: '已驳回' }
  return map[status] || status || '未知'
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.desc, .empty { margin-top: 10rpx; color: #667085; font-size: 26rpx; }
</style>
