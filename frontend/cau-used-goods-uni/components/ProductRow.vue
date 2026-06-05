<template>
  <view class="product">
    <image class="product-image" :src="product.image" mode="aspectFill" />
    <view class="product-body">
      <text class="product-title">{{ product.title }}</text>
      <text class="product-price">¥{{ product.price }}</text>
      <text class="product-meta">{{ product.meetLocation }}</text>
    </view>
    <slot />
  </view>
</template>

<script setup>
defineProps({
  product: { type: Object, required: true }
})
</script>

<style scoped lang="scss">
.product { display: flex; gap: 18rpx; align-items: center; }
.product-image { width: 148rpx; height: 116rpx; border-radius: 14rpx; background: #edf2ef; }
.product-body { display: flex; flex: 1; min-width: 0; flex-direction: column; gap: 8rpx; }
.product-title { overflow: hidden; color: #243129; font-size: 28rpx; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.product-price { color: #d46d32; font-size: 30rpx; font-weight: 700; }
.product-meta { color: #849089; font-size: 23rpx; }
</style>
