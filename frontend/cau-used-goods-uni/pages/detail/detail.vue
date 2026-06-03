<template>
  <view class="page">
    <image v-if="mainImage" class="goods-image" :src="mainImage" mode="aspectFill" />
    <view v-else class="goods-image placeholder">{{ categoryName }}</view>

    <view class="info-card">
      <view class="title">{{ product.title || '商品详情' }}</view>
      <view class="price">￥{{ product.price || 0 }}</view>
      <view class="meta">分类：{{ categoryName }}</view>
      <view class="meta">成色：{{ conditionText(product.conditionLevel) }}</view>
      <view class="meta">发布时间：{{ product.createTime || '暂无' }}</view>
      <view class="meta">交易地点：{{ product.meetLocation || '线下面交' }}</view>
    </view>

    <view class="info-card">
      <view class="section-title">商品描述</view>
      <view class="desc">{{ product.description || '暂无商品描述' }}</view>
    </view>

    <view class="info-card">
      <view class="section-title">卖家信息</view>
      <view class="meta">卖家编号：{{ product.sellerId || '-' }}</view>
      <view class="meta">联系方式需预约后在订单详情中查看</view>
    </view>

    <view class="button-row">
      <button class="plain-button" @click="favorite">收藏</button>
      <button class="main-button" @click="reserve">提交预约</button>
      <button class="danger-button" @click="report">举报</button>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { addFavorite, createOrder, createReport, getProductById } from '../../api/product'

const product = ref({})

const categoryNameMap = {
  1: '教材资料',
  2: '电子产品',
  3: '生活用品',
  4: '服饰鞋包',
  5: '运动户外',
  6: '其他'
}

const categoryName = computed(() => categoryNameMap[product.value?.categoryId] || '商品')
const mainImage = computed(() => product.value?.images?.[0] || '')

onLoad((query = {}) => {
  loadProduct(query.id)
})

const loadProduct = async (id) => {
  if (!id) {
    uni.showToast({ title: '商品不存在', icon: 'none' })
    return
  }
  try {
    product.value = await getProductById(id)
  } catch (error) {
    uni.showToast({ title: error.message || '商品加载失败', icon: 'none' })
  }
}

const conditionText = (level) => {
  const map = {
    NEW: '全新',
    LIKE_NEW: '九成新',
    GOOD: '八成新',
    FAIR: '七成新',
    OLD: '旧物'
  }
  return map[level] || level || '成色未填写'
}

const favorite = async () => {
  try {
    await addFavorite(product.value.id)
    uni.showToast({ title: '已收藏', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error.message || '收藏失败', icon: 'none' })
  }
}

const reserve = async () => {
  try {
    await createOrder({
      productId: product.value.id,
      meetTime: '',
      meetLocation: product.value.meetLocation || '线下面交',
      remark: ''
    })
    uni.showToast({ title: '预约已提交', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error.message || '预约失败', icon: 'none' })
  }
}

const report = async () => {
  try {
    await createReport({
      targetType: 'PRODUCT',
      targetId: product.value.id,
      reasonType: 'OTHER',
      description: '用户提交商品举报'
    })
    uni.showToast({ title: '举报已提交', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error.message || '举报失败', icon: 'none' })
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 32rpx;
  padding-bottom: 140rpx;
  background: #f6f7f9;
  box-sizing: border-box;
}

.goods-image {
  height: 420rpx;
  border-radius: 20rpx;
  background: #d1d5db;
}

.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #7b8794;
  font-size: 30rpx;
}

.info-card {
  margin-top: 24rpx;
  padding: 28rpx;
  border-radius: 16rpx;
  background: #ffffff;
}

.title {
  font-size: 38rpx;
  font-weight: 700;
  color: #1f2933;
}

.price {
  margin-top: 20rpx;
  font-size: 42rpx;
  font-weight: 700;
  color: #e11d48;
}

.meta {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #6b7280;
}

.section-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2933;
}

.desc {
  margin-top: 16rpx;
  line-height: 1.7;
  font-size: 28rpx;
  color: #374151;
}

.button-row {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  gap: 16rpx;
  padding: 20rpx 32rpx;
  background: #ffffff;
  box-sizing: border-box;
}

.plain-button,
.main-button,
.danger-button {
  flex: 1;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: 12rpx;
  font-size: 26rpx;
}

.plain-button {
  background: #f3f4f6;
  color: #374151;
}

.main-button {
  background: #1aad19;
  color: #ffffff;
}

.danger-button {
  background: #fee2e2;
  color: #dc2626;
}
</style>
