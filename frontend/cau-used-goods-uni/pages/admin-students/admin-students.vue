<template>
  <view class="page">
    <view class="section-title">学生认证审核</view>
    <view v-if="verifications.length === 0" class="empty">暂无待审核认证</view>
    <view v-for="item in verifications" :key="item.userId" class="card">
      <view class="name">{{ item.realName }} {{ item.studentId }}</view>
      <view class="desc">{{ item.college }}</view>
      <view class="actions">
        <button size="mini" class="pass" @click="review(item.userId, 'VERIFIED')">通过</button>
        <button size="mini" class="reject" @click="review(item.userId, 'REJECTED')">驳回</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getPendingStudentVerifications, reviewStudentVerification } from '../../api/admin'

const verifications = ref([])

const load = async () => {
  try {
    const result = await getPendingStudentVerifications()
    verifications.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  }
}

onShow(load)

const review = async (id, authStatus) => {
  try {
    await reviewStudentVerification(id, {
      authStatus,
      description: authStatus === 'VERIFIED' ? '学生认证审核通过' : '认证信息不符合要求'
    })
    uni.showToast({ title: '审核完成', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '审核失败', icon: 'none' })
  }
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.desc, .empty { margin-top: 8rpx; color: #667085; font-size: 26rpx; }
.actions { margin-top: 18rpx; display: flex; gap: 18rpx; }
.pass { background: #17a84b; color: #fff; }
.reject { background: #fff1f2; color: #ef4444; }
</style>
