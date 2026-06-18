<template>
  <view class="product">
    <image v-if="product.image" class="product-image" :src="product.image" mode="aspectFill" />
    <view v-else class="product-image product-placeholder">商品</view>
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
.product {
  display: flex;
  align-items: center;
  gap: 18rpx;
  width: 100%;
  max-width: 100%;
  overflow: hidden;
}

.product-image {
  display: flex;
  flex: 0 0 148rpx;
  width: 148rpx;
  height: 116rpx;
  align-items: center;
  justify-content: center;
  border-radius: 14rpx;
  background: #edf2ef;
}

.product-placeholder {
  color: #9aa5a1;
  font-size: 22rpx;
}

.product-body {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 8rpx;
}

.product-title {
  overflow: hidden;
  color: #243129;
  font-size: 28rpx;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-price {
  color: #d46d32;
  font-size: 30rpx;
  font-weight: 700;
}

.product-meta {
  overflow: hidden;
  color: #849089;
  font-size: 23rpx;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
