<template>
  <view class="page">
    <view class="header">
      <text class="title">系统消息</text>
      <text class="subtitle">订单进度、举报处理和平台通知</text>
    </view>

    <view v-if="messages.length" class="list">
      <view v-for="item in messages" :key="item.id" class="card" @click="open(item)">
        <view class="dot" :class="{ read: item.read }" />
        <view class="body">
          <view class="head">
            <text class="message-title">{{ item.title }}</text>
            <text class="time">{{ item.createdAt }}</text>
          </view>
          <text class="content">{{ item.content }}</text>
        </view>
      </view>
    </view>

    <EmptyState v-else title="暂无系统消息" detail="有新的交易进度或平台通知时会显示在这里" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import EmptyState from '../../components/EmptyState.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError } from '../../utils/navigation'

const SYSTEM_TYPES = ['ORDER_CREATED', 'ORDER_CONFIRMED', 'ORDER_CANCELED', 'ORDER_TIMEOUT', 'REPORT_HANDLED', 'SYSTEM_NOTICE']
const messages = ref([])

onShow(async () => {
  try {
    const list = await tradeService.getMessages()
    messages.value = list.filter((item) => SYSTEM_TYPES.includes(item.type || item.messageType))
  } catch (error) {
    showError(error)
  }
})

function open(item) {
  navigate('/pages/interaction/message-detail', { id: item.id })
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f7; box-sizing: border-box; }
.header { padding: 18rpx 6rpx 26rpx; }
.title, .subtitle { display: block; }
.title { color: #202124; font-size: 40rpx; font-weight: 800; }
.subtitle { margin-top: 8rpx; color: #8a8f94; font-size: 24rpx; }
.list { display: flex; flex-direction: column; gap: 16rpx; }
.card { display: flex; gap: 16rpx; padding: 24rpx; border-radius: 22rpx; background: #fff; box-shadow: 0 8rpx 28rpx rgba(23, 33, 43, .04); }
.dot { width: 16rpx; height: 16rpx; margin-top: 12rpx; flex: 0 0 16rpx; border-radius: 50%; background: #f04444; }
.dot.read { background: #c9d0d6; }
.body { flex: 1; min-width: 0; }
.head { display: flex; justify-content: space-between; gap: 16rpx; }
.message-title { flex: 1; min-width: 0; overflow: hidden; color: #222; font-size: 30rpx; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.time { flex-shrink: 0; color: #a0a6ad; font-size: 22rpx; }
.content { display: block; margin-top: 12rpx; color: #70777f; font-size: 25rpx; line-height: 1.6; }
</style>
