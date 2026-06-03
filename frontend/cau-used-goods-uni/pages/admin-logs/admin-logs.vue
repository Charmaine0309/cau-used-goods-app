<template>
  <view class="page">
    <view class="section-title">管理员日志</view>
    <view v-if="logs.length === 0" class="empty">暂无日志</view>
    <view v-for="log in logs" :key="log.id" class="log-item">
      <view class="name">{{ log.operationType || '操作' }}</view>
      <view class="desc">{{ log.targetType || '对象' }} · {{ log.createTime || '' }}</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAdminLogs } from '../../api/admin'

const logs = ref([])

const load = async () => {
  try {
    const result = await getAdminLogs()
    logs.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '日志加载失败', icon: 'none' })
  }
}

onShow(load)
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.log-item, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.desc, .empty { margin-top: 8rpx; color: #667085; font-size: 26rpx; }
</style>
