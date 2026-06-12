<template>
  <view class="page">
    <view class="hero">
      <text class="eyebrow">CAU CAMPUS MARKET</text>
      <text class="hero-title">消息中心</text>
      <text class="hero-copy">订单进度、举报处理和系统通知会在这里提醒你</text>
    </view>

    <view v-if="messages.length" class="message-panel">
      <view class="panel-head">
        <text class="panel-title">最新消息</text>
        <text class="panel-count">{{ messages.length }} 条</text>
      </view>
      <view v-for="message in messages" :key="message.id" class="card message" @click="open(message.id)">
        <view class="dot" :class="{ read: message.read }" />
        <view class="body">
          <view class="head">
            <text class="message-title">{{ message.title }}</text>
            <text class="time">{{ message.createdAt }}</text>
          </view>
          <text class="content">{{ message.content }}</text>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>
    <view v-else class="empty-wrap">
      <EmptyState title="暂无消息" detail="订单进度与举报处理结果会在这里提醒你" />
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import EmptyState from '../../components/EmptyState.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError } from '../../utils/navigation'

const messages = ref([])
onShow(async () => {
  try { messages.value = await tradeService.getMessages() } catch (error) { showError(error) }
})
const open = (id) => navigate('/pages/interaction/message-detail', { id })
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 0 26rpx 42rpx;
  background: #f3f8f5;
  box-sizing: border-box;
}

.hero {
  margin: 0 -26rpx;
  padding: 70rpx 34rpx 86rpx;
  border-radius: 0 0 42rpx 42rpx;
  background: linear-gradient(145deg, #23734f, #2f8b62);
  color: #fff;
}

.eyebrow,
.hero-title,
.hero-copy {
  display: block;
}

.eyebrow {
  color: rgba(255, 255, 255, .72);
  font-size: 20rpx;
  letter-spacing: 3rpx;
}

.hero-title {
  margin-top: 14rpx;
  font-size: 42rpx;
  line-height: 50rpx;
  font-weight: 700;
}

.hero-copy {
  margin-top: 14rpx;
  color: rgba(255, 255, 255, .78);
  font-size: 24rpx;
  line-height: 1.5;
}

.message-panel {
  margin-top: -48rpx;
  padding: 26rpx 0 0;
  border-radius: 28rpx;
  background: #fff;
  box-shadow: 0 14rpx 34rpx rgba(31, 106, 73, .06);
  overflow: hidden;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28rpx 22rpx;
}

.panel-title {
  color: #20352b;
  font-size: 32rpx;
  font-weight: 700;
}

.panel-count {
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  background: #eef7f0;
  color: #17a84b;
  font-size: 22rpx;
}

.card {
  padding: 24rpx 28rpx;
  background: #fff;
}

.message {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
  border-top: 1rpx solid #eef0f3;
}

.dot {
  width: 16rpx;
  height: 16rpx;
  margin-top: 14rpx;
  border-radius: 50%;
  background: #17a84b;
  flex-shrink: 0;
}

.dot.read {
  background: #ccd4d0;
}

.body {
  flex: 1;
  min-width: 0;
}

.head {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
}

.message-title {
  color: #20352b;
  font-size: 30rpx;
  font-weight: 700;
  line-height: 38rpx;
}

.time {
  flex-shrink: 0;
  color: #98a39d;
  font-size: 22rpx;
}

.content {
  display: block;
  overflow: hidden;
  margin-top: 10rpx;
  color: #738077;
  font-size: 25rpx;
  line-height: 36rpx;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.arrow {
  margin-top: 22rpx;
  color: #b2bdca;
  font-size: 38rpx;
}

.empty-wrap {
  margin-top: -48rpx;
  border-radius: 28rpx;
  background: #fff;
  box-shadow: 0 14rpx 34rpx rgba(31, 106, 73, .06);
  overflow: hidden;
}
</style>
