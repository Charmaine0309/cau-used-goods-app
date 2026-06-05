<template>
  <view class="page">
    <view v-if="products.length">
      <view v-for="product in products" :key="product.id" class="card">
        <ProductRow :product="product" />
        <view class="favorite-actions">
          <button class="btn btn-primary" :disabled="product.status !== 'ON_SALE'" @click="appointment(product.id)">
            {{ product.status === 'ON_SALE' ? '提交预约' : '当前不可预约' }}
          </button>
          <button class="btn btn-plain" @click="remove(product.id)">取消收藏</button>
        </view>
      </view>
    </view>
    <EmptyState v-else title="收藏夹空空的" detail="在商品详情页点击收藏后，会显示在这里" />
  </view>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import ProductRow from '../../components/ProductRow.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError, showSuccess } from '../../utils/navigation'

const products = ref([])
onShow(load)

async function load() {
  try {
    products.value = await tradeService.getFavorites()
  } catch (error) {
    showError(error)
  }
}

function appointment(productId) {
  navigate('/pages/order/appointment', { productId })
}

async function remove(productId) {
  try {
    await tradeService.removeFavorite(productId)
    showSuccess('已取消收藏')
    load()
  } catch (error) {
    showError(error)
  }
}
</script>

<style scoped lang="scss">
.favorite-actions { display: flex; gap: 14rpx; margin-top: 22rpx; }
.favorite-actions .btn { flex: 1; min-height: 68rpx; font-size: 25rpx; }
</style>
