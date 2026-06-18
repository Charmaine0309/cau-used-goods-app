<template>
  <view v-if="message" class="page">
    <view class="hero">
      <text class="eyebrow">SYSTEM MESSAGE</text>
      <text class="hero-title">消息详情</text>
      <text class="hero-copy">查看订单进度、举报处理和平台通知的完整内容</text>
    </view>

    <view class="card">
      <view class="card-head">
        <view class="message-icon">系</view>
        <view class="head-body">
          <text class="title">{{ message.title }}</text>
          <text class="meta">{{ message.createdAt }}</text>
        </view>
      </view>
      <text class="content">{{ message.content }}</text>
    </view>

    <view v-if="message.targetType === 'ORDER'" class="action-card">
      <view>
        <text class="action-title">相关订单</text>
        <text class="action-desc">点击查看该消息对应的订单详情</text>
      </view>
      <button class="btn btn-primary" @click="openOrder">查看订单</button>
    </view>
  </view>
</template>

<script setup>
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { tradeService } from '../../services/trade'
import { navigate, showError } from '../../utils/navigation'

const message = ref()

onLoad(async (options) => {
  try {
    message.value = await tradeService.getMessage(options.id)
    await tradeService.markMessageRead(options.id)
  } catch (error) {
    showError(error)
  }
})

function openOrder() {
  navigate('/pages/order/detail', { id: message.value.targetId })
}
</script>

<style scoped lang="scss">
.page { min-height: 100vh; padding: 30rpx 28rpx 48rpx; background: #f5f8f6; box-sizing: border-box; }
.hero { padding: 34rpx 30rpx; border-radius: 28rpx; background: linear-gradient(135deg, #23734f, #3e9b72); color: #fff; box-shadow: 0 12rpx 32rpx rgba(35, 115, 79, .18); }
.eyebrow, .hero-title, .hero-copy, .title, .meta, .content, .action-title, .action-desc { display: block; }
.eyebrow { color: rgba(255,255,255,.72); font-size: 20rpx; letter-spacing: 3rpx; }
.hero-title { margin-top: 12rpx; font-size: 40rpx; font-weight: 800; }
.hero-copy { margin-top: 10rpx; color: rgba(255,255,255,.78); font-size: 24rpx; line-height: 1.5; }
.card { display: flex; flex-direction: column; gap: 28rpx; margin-top: 24rpx; padding: 32rpx 30rpx; border-radius: 28rpx; background: #fff; box-shadow: 0 10rpx 30rpx rgba(28, 68, 52, .06); }
.card-head { display: flex; align-items: center; gap: 20rpx; }
.message-icon { display: flex; width: 72rpx; height: 72rpx; flex-shrink: 0; align-items: center; justify-content: center; border-radius: 50%; background: #e7f4ec; color: #23734f; font-size: 28rpx; font-weight: 800; }
.head-body { flex: 1; min-width: 0; }
.title { color: #26342f; font-size: 34rpx; font-weight: 800; line-height: 1.45; }
.meta { margin-top: 8rpx; color: #98a39d; font-size: 23rpx; }
.content { color: #425148; font-size: 28rpx; line-height: 1.9; white-space: pre-line; }
.action-card { display: flex; align-items: center; justify-content: space-between; gap: 20rpx; margin-top: 24rpx; padding: 28rpx 30rpx; border-radius: 28rpx; background: #fff; box-shadow: 0 10rpx 30rpx rgba(28, 68, 52, .06); }
.action-title { color: #26342f; font-size: 29rpx; font-weight: 800; }
.action-desc { margin-top: 8rpx; color: #89938f; font-size: 23rpx; }
.btn-primary { flex-shrink: 0; width: 180rpx; height: 68rpx; border-radius: 999rpx; background: #23734f; color: #fff; font-size: 25rpx; line-height: 68rpx; }
</style>
