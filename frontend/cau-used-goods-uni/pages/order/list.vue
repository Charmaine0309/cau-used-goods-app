<template>
  <view class="page">
    <view class="tabs">
      <view :class="{ active: role === 'buyer' }" @click="switchRole('buyer')">我买到的</view>
      <view :class="{ active: role === 'seller' }" @click="switchRole('seller')">我卖出的</view>
    </view>
    <view v-if="orders.length">
      <OrderCard v-for="order in orders" :key="order.id" :order="order" :role="role" @open="openOrder" />
    </view>
    <EmptyState v-else title="暂无订单" detail="订单会按买家与卖家身份分别展示" />
  </view>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import OrderCard from '../../components/OrderCard.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError } from '../../utils/navigation'

const role = ref('buyer')
const orders = ref([])

onShow(load)

async function load() {
  try {
    orders.value = await tradeService.getOrders(role.value)
  } catch (error) {
    showError(error)
  }
}

function switchRole(value) {
  role.value = value
  load()
}

function openOrder(id) {
  navigate('/pages/order/detail', { id })
}
</script>

<style scoped lang="scss">
.tabs { display: flex; margin: 8rpx 0 24rpx; padding: 8rpx; border-radius: 40rpx; background: #e9efeb; }
.tabs view { flex: 1; padding: 16rpx; border-radius: 32rpx; color: #738077; text-align: center; font-size: 27rpx; }
.tabs .active { color: #2f6b4f; background: #fff; font-weight: 700; box-shadow: 0 4rpx 14rpx rgba(36, 49, 41, .06); }
</style>
