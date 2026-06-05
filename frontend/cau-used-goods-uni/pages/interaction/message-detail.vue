<template>
  <view v-if="message" class="page">
    <view class="card">
      <text class="title">{{ message.title }}</text>
      <text class="meta">{{ message.createdAt }}</text>
      <text class="content">{{ message.content }}</text>
    </view>
    <button v-if="message.targetType === 'ORDER'" class="btn btn-primary" @click="openOrder">查看相关订单</button>
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
.card { display: flex; flex-direction: column; gap: 18rpx; }
.title { font-size: 36rpx; font-weight: 700; }
.meta { color: #98a39d; font-size: 23rpx; }
.content { color: #425148; font-size: 28rpx; line-height: 1.8; }
</style>
