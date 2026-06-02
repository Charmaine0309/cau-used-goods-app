<template>
  <view class="card order-card" @click="$emit('open', order.id)">
    <view class="order-head">
      <text class="order-no">订单 {{ order.id }}</text>
      <StatusBadge :label="status.label" :tone="status.tone" />
    </view>
    <ProductRow :product="order.product" />
    <view class="order-foot">
      <text>{{ role === 'seller' ? `买家：${order.buyerName}` : `卖家：${order.sellerName}` }}</text>
      <text>{{ order.createdAt }}</text>
    </view>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import { ORDER_STATUS } from '../utils/constants'
import ProductRow from './ProductRow.vue'
import StatusBadge from './StatusBadge.vue'

const props = defineProps({
  order: { type: Object, required: true },
  role: { type: String, default: 'buyer' }
})
defineEmits(['open'])

const status = computed(() => ORDER_STATUS[props.order.status] || { label: props.order.status, tone: 'muted' })
</script>

<style scoped lang="scss">
.order-card { padding-bottom: 20rpx; }
.order-head, .order-foot { display: flex; align-items: center; justify-content: space-between; }
.order-head { margin-bottom: 20rpx; }
.order-no { color: #738077; font-size: 23rpx; }
.order-foot { margin-top: 20rpx; color: #849089; font-size: 22rpx; }
</style>
