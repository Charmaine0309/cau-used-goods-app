<template>
  <view v-if="product" class="page">
    <swiper v-if="product.images && product.images.length" class="gallery" indicator-dots circular>
      <swiper-item v-for="image in product.images" :key="image">
        <image class="gallery-image" :src="image" mode="aspectFill" />
      </swiper-item>
    </swiper>
    <view v-else class="gallery placeholder">暂无图片</view>

    <view class="card">
      <view class="price-line">
        <text class="price">¥{{ product.priceText }}</text>
        <text class="status">{{ statusText }}</text>
      </view>
      <view class="title">{{ product.title }}</view>
      <view class="meta">{{ product.conditionText }} · {{ product.viewCount || 0 }} 次浏览 · {{ product.timeText }}</view>
    </view>

    <view class="card">
      <view class="section-title">商品描述</view>
      <view class="description">{{ product.description || '卖家暂未填写描述' }}</view>
    </view>

    <view class="card">
      <view class="section-title">面交信息</view>
      <view class="description">建议地点：{{ product.meetLocation || '预约后协商' }}</view>
      <view class="privacy">为保护隐私，联系方式仅在预约进入待面交后向交易双方展示。</view>
    </view>

    <view class="card seller">
      <view class="avatar">{{ sellerAvatarText }}</view>
      <view>
        <view class="seller-name">{{ sellerName }}</view>
        <view class="meta">{{ product.seller?.college || '中国农业大学' }}</view>
      </view>
    </view>

    <view class="bottom">
      <button class="minor" @click="toggleFavorite">{{ favoriteText }}</button>
      <button class="minor report" @click="report">举报</button>
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
import { buildCategoryMap, formatProduct, getStatusText } from '../../utils/product-format'
import { getToken, isVerifiedUser } from '../../utils/auth'
import { navigate } from '../../utils/navigation'
import { displayUserName, isBannedUserStatus, isCanceledUserStatus } from '../../utils/user-format'

const product = ref(null)
const isFavorite = ref(false)

const statusText = computed(() => getStatusText(product.value?.status))
const favoriteText = computed(() => isFavorite.value ? '已收藏' : '收藏')
const sellerName = computed(() => displayUserName(product.value?.seller || {}, 'CAU 同学'))
const sellerAvatarText = computed(() => {
  const status = product.value?.seller?.accountStatus || product.value?.seller?.account_status || product.value?.seller?.status
  if (isBannedUserStatus(status) || isCanceledUserStatus(status)) return '账'
  return (sellerName.value || '卖').slice(0, 1)
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

const toggleFavorite = async () => {
  if (!ensureVerified()) return

  try {
    if (isFavorite.value) {
      await removeFavorite(product.value.id)
    } else {
      await addFavorite(product.value.id)
    }

    isFavorite.value = !isFavorite.value
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

    if (getToken()) {
      isFavorite.value = (await checkFavorite(id)).favorited
    }
  } catch (error) {
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
  margin-top: 14rpx;
  color: #89938f;
  font-size: 23rpx;
}

.section-title,
.seller-name {
  font-weight: 700;
}

.description,
.privacy {
  margin-top: 16rpx;
  color: #58645f;
  line-height: 1.7;
}

.privacy {
  color: #9a7745;
  font-size: 23rpx;
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
  gap: 14rpx;
  padding: 16rpx 20rpx calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
}

.minor,
.primary {
  height: 74rpx;
  border-radius: 999rpx;
  font-size: 26rpx;
  line-height: 74rpx;
}

.minor {
  width: 132rpx;
  background: #edf4f1;
  color: #23734f;
}

.report {
  color: #b85d45;
}

.primary {
  flex: 1;
  background: #23734f;
  color: #fff;
}

.primary[disabled] {
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
