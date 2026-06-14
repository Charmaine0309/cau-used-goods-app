<template>
  <view v-if="product" class="page">
    <swiper v-if="visibleImages.length" class="gallery" indicator-dots circular>
      <swiper-item v-for="image in visibleImages" :key="image">
        <image class="gallery-image" :src="image" mode="aspectFill" @error="markImageFailed(image)" />
      </swiper-item>
    </swiper>
    <view v-else class="gallery placeholder">图片未找到</view>

    <view class="card">
      <view class="price-line">
        <text class="price">￥{{ product.priceText }}</text>
        <text class="status">{{ statusText }}</text>
      </view>
      <view class="title">{{ product.title }}</view>
      <view class="meta">
        <text>{{ product.conditionText }}</text>
        <text>{{ product.viewCount || 0 }} 浏览</text>
        <text>{{ favoriteCount }} 收藏</text>
        <text>{{ product.timeText }}</text>
      </view>
    </view>

    <view class="card">
      <view class="section-title">商品描述</view>
      <view class="description">{{ product.description || '卖家暂未填写描述' }}</view>
    </view>

    <view class="card seller">
      <view class="avatar">{{ sellerAvatarText }}</view>
      <view>
        <view class="seller-name">{{ sellerName }}</view>
        <view class="meta single">{{ product.seller?.college || '中国农业大学' }}</view>
      </view>
    </view>

    <view class="bottom">
      <button class="icon-button favorite" :class="{ active: isFavorite }" @click="toggleFavorite">{{ isFavorite ? '★' : '☆' }}</button>
      <button class="icon-button report" @click="report">!</button>
      <button class="chat" :disabled="product.status !== 'ON_SALE'" @click="chat">聊一聊</button>
      <button class="primary" :disabled="product.status !== 'ON_SALE'" @click="reserve">
        {{ product.status === 'ON_SALE' ? '提交预约' : statusText }}
      </button>
    </view>
  </view>

  <view v-else class="page loading-page">
    <view class="load-text">正在加载商品详情...</view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import {
  addFavorite,
  checkFavorite,
  getProductById,
  listCategories,
  removeFavorite
} from '../../api/product'
import { createOrGetConversation } from '../../api/chat'
import { buildCategoryMap, formatProduct, getStatusText } from '../../utils/product-format'
import { getToken, getUser, isVerifiedUser } from '../../utils/auth'
import { navigate } from '../../utils/navigation'
import { displayUserName, isBannedUserStatus, isCanceledUserStatus } from '../../utils/user-format'

const product = ref(null)
const isFavorite = ref(false)
const failedImages = ref([])

const statusText = computed(() => getStatusText(product.value?.status))
const favoriteCount = computed(() => Number(product.value?.favoriteCount || product.value?.favorite_count || 0))
const visibleImages = computed(() => {
  const images = product.value?.images || []
  return images.filter((image) => !failedImages.value.includes(image))
})
const sellerName = computed(() => displayUserName(product.value?.seller || {}, 'CAU 同学'))
const sellerAvatarText = computed(() => {
  const status = product.value?.seller?.accountStatus || product.value?.seller?.account_status || product.value?.seller?.status
  if (isBannedUserStatus(status) || isCanceledUserStatus(status)) return '封'
  return (sellerName.value || '同').slice(0, 1)
})

const ensureVerified = () => {
  if (!getToken()) {
    uni.navigateTo({ url: '/pages/login/login' })
    return false
  }

  if (!isVerifiedUser()) {
    uni.navigateTo({ url: '/pages/student-auth/student-auth' })
    return false
  }

  return true
}

const adjustFavoriteCount = (delta) => {
  const current = favoriteCount.value
  product.value.favoriteCount = Math.max(0, current + delta)
}

const markImageFailed = (image) => {
  if (!failedImages.value.includes(image)) {
    failedImages.value = failedImages.value.concat(image)
  }
}

const toggleFavorite = async () => {
  if (!ensureVerified()) return

  try {
    if (isFavorite.value) {
      await removeFavorite(product.value.id)
      isFavorite.value = false
      adjustFavoriteCount(-1)
    } else {
      await addFavorite(product.value.id)
      isFavorite.value = true
      adjustFavoriteCount(1)
    }

    uni.showToast({ title: isFavorite.value ? '收藏成功' : '已取消收藏', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error.message || '收藏操作失败', icon: 'none' })
  }
}

const reserve = () => {
  if (!ensureVerified() || !product.value?.id) return
  navigate('/pages/order/appointment', { productId: product.value.id })
}

const report = () => {
  if (!ensureVerified() || !product.value?.id) return
  navigate('/pages/interaction/report', { targetType: 'PRODUCT', targetId: product.value.id })
}

const chat = async () => {
  if (!ensureVerified() || !product.value?.id) return
  try {
    const conversation = await createOrGetConversation(product.value.id)
    const currentUserId = Number(getUser()?.id || getUser()?.userId || 0)
    const targetUserId = Number(conversation.buyerId) === currentUserId ? conversation.sellerId : conversation.buyerId
    navigate('/pages/chat/chat', {
      conversationId: conversation.id,
      title: product.value.title,
      targetUserId
    })
  } catch (error) {
    uni.showToast({ title: error.message || '暂时无法发起私信', icon: 'none' })
  }
}

onLoad(async ({ id }) => {
  if (!id) {
    uni.showToast({ title: '商品不存在', icon: 'none' })
    return
  }

  try {
    const [detail, categories] = await Promise.all([
      getProductById(id),
      listCategories()
    ])

    product.value = formatProduct(detail, buildCategoryMap(categories))
    uni.setStorageSync(`product-detail-cache-${id}`, product.value)
    failedImages.value = []

    if (getToken()) {
      isFavorite.value = (await checkFavorite(id)).favorited
    }
  } catch (error) {
    const cached = uni.getStorageSync(`product-detail-cache-${id}`)
    if (cached) {
      product.value = cached
      failedImages.value = []
      uni.showToast({ title: '商品已被预约，显示最近一次详情', icon: 'none' })
      return
    }
    uni.showToast({ title: error.message || '商品加载失败', icon: 'none' })
  }
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding-bottom: 130rpx;
}

.gallery,
.gallery-image {
  width: 100%;
  height: 600rpx;
  background: #e8efeb;
}

.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9aa5a1;
  font-size: 28rpx;
}

.card {
  margin: 20rpx;
  padding: 24rpx;
  border-radius: 18rpx;
  background: #fff;
}

.price-line,
.seller {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.price {
  color: #e36a3e;
  font-size: 46rpx;
  font-weight: 700;
}

.status {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #e7f4ec;
  color: #23734f;
  font-size: 23rpx;
}

.title {
  margin-top: 14rpx;
  font-size: 36rpx;
  font-weight: 700;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx 18rpx;
  margin-top: 14rpx;
  color: #89938f;
  font-size: 23rpx;
}

.meta.single {
  display: block;
}

.section-title,
.seller-name {
  font-weight: 700;
}

.description {
  margin-top: 16rpx;
  color: #58645f;
  line-height: 1.7;
}

.seller {
  justify-content: flex-start;
}

.avatar {
  display: flex;
  width: 80rpx;
  height: 80rpx;
  margin-right: 16rpx;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #e7f4ec;
  color: #23734f;
  font-weight: 700;
}

.bottom {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  gap: 12rpx;
  padding: 16rpx 20rpx calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
}

.icon-button,
.chat,
.primary {
  height: 74rpx;
  border-radius: 999rpx;
  font-size: 26rpx;
  line-height: 74rpx;
}

.icon-button {
  width: 74rpx;
  padding: 0;
  background: #edf4f1;
  color: #23734f;
  font-size: 34rpx;
}

.favorite.active { color: #f5b301; }
.report { color: #d64545; font-weight: 800; }

.chat {
  width: 138rpx;
  background: #fff6e9;
  color: #9a6a1d;
}

.primary {
  flex: 1;
  background: #23734f;
  color: #fff;
}

.primary[disabled],
.chat[disabled] {
  background: #b8c5c0;
  color: #fff;
}

.loading-page {
  display: flex;
  align-items: center;
  justify-content: center;
}

.load-text {
  color: #929c98;
  font-size: 26rpx;
}
</style>
