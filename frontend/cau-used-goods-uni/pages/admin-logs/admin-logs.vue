<template>
  <view class="page">
    <view class="section-title">管理员日志</view>
    <view v-if="logs.length === 0" class="empty">暂无日志</view>
    <view v-for="log in logs" :key="log.id" class="log-item" @click="goRelatedPage(log)">
      <view class="log-main">
        <view>
          <view class="name">{{ operationLabel(log.operationType) }}</view>
          <view class="desc">{{ formatDateTime(log.createTime) }}</view>
        </view>
        <text class="arrow">›</text>
      </view>
      <view v-if="log.description" class="detail">{{ log.description }}</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAdminLogs } from '../../api/admin'

const logs = ref([])

const operationMap = {
  USER_DISABLE: '禁用用户',
  USER_ENABLE: '启用用户',
  PRODUCT_OFF_SHELF: '下架商品',
  PRODUCT_ON_SALE: '上架商品',
  REPORT_RESOLVE: '处理举报',
  REPORT_REJECT: '驳回举报',
  REPORT_CLOSE: '关闭举报',
  NOTICE_PUBLISH: '发布公告',
  NOTICE_OFFLINE: '下线公告',
  WORD_CREATE: '新增敏感词',
  WORD_DISABLE: '禁用敏感词',
  CATEGORY_CREATE: '新增标签',
  CATEGORY_UPDATE: '编辑标签',
  CATEGORY_ENABLE: '启用标签',
  CATEGORY_DISABLE: '停用标签',
  ORDER_EXCEPTION_CLOSE: '异常关闭订单',
  CREATE_NOTICE: '新增公告',
  UPDATE_NOTICE: '编辑公告',
  CREATE_WORD: '新增敏感词',
  UPDATE_WORD: '编辑敏感词',
  DELETE_WORD: '禁用敏感词',
  HANDLE_APPEAL: '处理申诉',
  STUDENT_VERIFY_APPROVE: '通过学生认证',
  STUDENT_VERIFY_REJECT: '驳回学生认证'
}

const load = async () => {
  try {
    const result = await getAdminLogs()
    logs.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '日志加载失败', icon: 'none' })
  }
}

const operationLabel = (value) => operationMap[value] || value || '操作'

const pad = (value) => String(value).padStart(2, '0')

const formatDateTime = (value) => {
  if (!value) return ''
  const date = new Date(value)
  if (!Number.isNaN(date.getTime())) {
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
  }
  return String(value).replace('T', ' ').replace(/\+\d{2}:\d{2}$/, '')
}

const relatedPage = (log) => {
  if (!log) return ''
  if (log.operationType === 'STUDENT_VERIFY_APPROVE' || log.operationType === 'STUDENT_VERIFY_REJECT') {
    return '/pages/admin-students/admin-students'
  }
  const map = {
    USER: '/pages/admin-users/admin-users',
    PRODUCT: '/pages/admin-products/admin-products',
    REPORT: '/pages/admin-reports/admin-reports',
    APPEAL: '/pages/admin-reports/admin-reports',
    NOTICE: '/pages/admin-announcements/admin-announcements',
    WORD: '/pages/admin-sensitive/admin-sensitive',
    CATEGORY: '/pages/admin-categories/admin-categories'
  }
  return map[log.targetType] || ''
}

const goRelatedPage = (log) => {
  const url = relatedPage(log)
  if (!url) {
    uni.showToast({ title: '暂无对应管理页面', icon: 'none' })
    return
  }
  uni.navigateTo({ url })
}

onShow(load)
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f5f6f8;
  box-sizing: border-box;
}

.section-title {
  margin: 24rpx 0 18rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2933;
}

.log-item,
.empty {
  padding: 28rpx;
  border-radius: 16rpx;
  background: #fff;
  margin-bottom: 18rpx;
}

.log-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
}

.name {
  font-size: 30rpx;
  font-weight: 700;
  color: #1f2933;
}

.desc,
.empty,
.detail {
  margin-top: 8rpx;
  color: #667085;
  font-size: 26rpx;
}

.detail {
  line-height: 38rpx;
  color: #8a96a8;
}

.arrow {
  color: #b2bdca;
  font-size: 42rpx;
  line-height: 42rpx;
}
</style>
