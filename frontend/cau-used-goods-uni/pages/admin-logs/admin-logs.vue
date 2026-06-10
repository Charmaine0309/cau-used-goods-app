<template>
  <view class="page">
    <scroll-view class="filter-row" scroll-x enhanced :show-scrollbar="false">
      <view class="filter-inner">
        <view
          v-for="item in logFilters"
          :key="item.value"
          :class="['filter-chip', activeFilter === item.value ? 'active' : '']"
          @tap.stop="selectFilter(item.value)"
        >
          {{ item.label }} {{ countByFilter(item.value) }}
        </view>
      </view>
    </scroll-view>

    <view v-if="filteredLogs.length === 0" class="empty">暂无日志</view>
    <view v-for="log in filteredLogs" :key="log.id" class="log-item" @click="goRelatedPage(log)">
      <view class="log-main">
        <view>
          <view class="name">{{ operationLabel(log.operationType, log) }}</view>
          <view class="desc">{{ formatDateTime(log.createTime) }}</view>
        </view>
        <text class="arrow">›</text>
      </view>
      <view v-if="formatDescription(log)" class="detail">{{ formatDescription(log) }}</view>
      <view v-if="sourceText(log)" class="source">{{ sourceText(log) }}</view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAdminLogs } from '../../api/admin'

const logs = ref([])
const activeFilter = ref('ALL')

const logFilters = [
  { label: '全部', value: 'ALL' },
  { label: '商品', value: 'PRODUCT' },
  { label: '用户', value: 'USER' },
  { label: '订单', value: 'ORDER' },
  { label: '举报', value: 'REPORT' },
  { label: '申诉', value: 'APPEAL' },
  { label: '公告', value: 'NOTICE' },
  { label: '敏感词', value: 'WORD' },
  { label: '标签', value: 'CATEGORY' }
]

const operationMap = {
  USER_DISABLE: '禁用用户',
  USER_ENABLE: '启用用户',
  PRODUCT_OFF_SHELF: '下架商品',
  PRODUCT_ON_SALE: '上架商品',
  REPORT_RESOLVE: '处理举报',
  REPORT_APPROVE: '通过举报',
  REPORT_HANDLE: '处理举报',
  REPORT_REJECT: '驳回举报',
  REPORT_CLOSE: '关闭举报',
  HANDLE_APPEAL: '处理申诉',
  NOTICE_PUBLISH: '发布公告',
  NOTICE_OFFLINE: '下线公告',
  STATUS_NOTICE: '更新公告状态',
  CREATE_NOTICE: '新增公告',
  UPDATE_NOTICE: '编辑公告',
  DELETE_NOTICE: '删除公告',
  WORD_CREATE: '新增敏感词',
  WORD_DISABLE: '禁用敏感词',
  CREATE_WORD: '新增敏感词',
  UPDATE_WORD: '编辑敏感词',
  DELETE_WORD: '禁用敏感词',
  CATEGORY_CREATE: '新增标签',
  CATEGORY_UPDATE: '编辑标签',
  CATEGORY_ENABLE: '启用标签',
  CATEGORY_DISABLE: '停用标签',
  ORDER_EXCEPTION_CLOSE: '异常关闭订单',
  STUDENT_VERIFY_APPROVE: '通过学生认证',
  STUDENT_VERIFY_REJECT: '驳回学生认证'
}

const statusLabelMap = {
  DRAFT: '草稿',
  PUBLISHED: '已发布',
  OFFLINE: '已下线',
  PROCESSING: '处理中',
  APPROVED: '已通过',
  RESOLVED: '已处理',
  REJECTED: '已驳回',
  CLOSED: '已关闭',
  ENABLED: '已启用',
  DISABLED: '已停用',
  NORMAL: '正常',
  HIDDEN: '已隐藏',
  DELETED: '已删除'
}

const targetLabelMap = {
  USER: '用户',
  PRODUCT: '商品',
  REPORT: '举报',
  APPEAL: '申诉',
  NOTICE: '公告',
  WORD: '敏感词',
  CATEGORY: '标签',
  ORDER: '订单'
}

const descriptionMap = [
  [/^update announcement status:\s*(.+)$/i, (_, status) => announcementStatusDescription(status)],
  [/^create announcement:\s*(.+)$/i, (_, title) => `新增公告：${title}`],
  [/^disable sensitive word by delete operation$/i, () => '禁用敏感词'],
  [/^create sensitive word:\s*(.+)$/i, (_, word) => `新增敏感词：${word}`],
  [/^update sensitive word:\s*(.+)$/i, (_, word) => `编辑敏感词：${word}`],
  [/^mark report as processing$/i, () => '举报标记为处理中'],
  [/^handle report.*APPROVED$/i, () => '举报已通过'],
  [/^handle report.*REJECTED$/i, () => '举报已驳回'],
  [/^mark appeal #?(\d+)? as PROCESSING$/i, (_, id) => `申诉${id ? ` #${id}` : ''}标记为处理中`],
  [/^handle appeal #?(\d+)?:\s*APPROVED$/i, (_, id) => `申诉 #${id}已通过`],
  [/^handle appeal #?(\d+)?:\s*REJECTED$/i, (_, id) => `申诉 #${id}已驳回`]
]

const logCategory = (log) => {
  if (!log) return ''
  if (['STUDENT_VERIFY_APPROVE', 'STUDENT_VERIFY_REJECT', 'USER_DISABLE', 'USER_ENABLE'].includes(log.operationType)) return 'USER'
  if (log.operationType?.includes('NOTICE')) return 'NOTICE'
  if (log.operationType?.includes('WORD')) return 'WORD'
  if (log.operationType?.includes('CATEGORY')) return 'CATEGORY'
  return log.targetType || ''
}

const filteredLogs = computed(() => {
  if (activeFilter.value === 'ALL') return logs.value
  return logs.value.filter((log) => logCategory(log) === activeFilter.value)
})

const countByFilter = (value) => {
  if (value === 'ALL') return logs.value.length
  return logs.value.filter((log) => logCategory(log) === value).length
}

const selectFilter = (value) => {
  activeFilter.value = value
}

const load = async () => {
  try {
    const result = await getAdminLogs()
    logs.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '日志加载失败', icon: 'none' })
  }
}

const operationLabel = (value, log = null) => {
  if (value === 'STATUS_NOTICE') {
    const raw = String(log?.description || '')
    if (raw.includes('PUBLISHED') || raw.includes('已发布')) return '发布公告'
    if (raw.includes('OFFLINE') || raw.includes('已下线')) return '下线公告'
    if (raw.includes('DRAFT') || raw.includes('草稿')) return '转为草稿'
  }
  return operationMap[value] || translateOperation(value)
}

const translateOperation = (value) => {
  if (!value) return '操作'
  return String(value)
    .replaceAll('STATUS', '更新状态')
    .replaceAll('CREATE', '新增')
    .replaceAll('UPDATE', '编辑')
    .replaceAll('DELETE', '删除')
    .replaceAll('NOTICE', '公告')
    .replaceAll('WORD', '敏感词')
    .replaceAll('REPORT', '举报')
    .replaceAll('APPEAL', '申诉')
    .replaceAll('PRODUCT', '商品')
    .replaceAll('USER', '用户')
    .replaceAll('_', '')
}

const translateStatus = (status) => statusLabelMap[String(status || '').trim()] || status || ''

const replaceKnownWords = (value) => {
  let text = String(value || '')
  Object.keys(statusLabelMap).forEach((key) => {
    text = text.replaceAll(key, statusLabelMap[key])
  })
  Object.keys(targetLabelMap).forEach((key) => {
    text = text.replaceAll(key, targetLabelMap[key])
  })
  return text
}

const formatDescription = (log) => {
  const raw = String(log?.description || '').trim()
  const target = targetLabelMap[log?.targetType] || '对象'
  const idText = log?.targetId ? ` #${log.targetId}` : ''

  for (const [pattern, formatter] of descriptionMap) {
    const matched = raw.match(pattern)
    if (matched) return formatter(...matched)
  }

  const replaced = replaceKnownWords(raw)
  if (replaced && !/[A-Za-z_]/.test(replaced)) return replaced

  switch (log?.operationType) {
    case 'STATUS_NOTICE':
      return replaced || `公告${idText}状态已更新`
    case 'REPORT_HANDLE':
      return `举报${idText}标记为处理中`
    case 'REPORT_APPROVE':
      return `举报${idText}已通过`
    case 'REPORT_REJECT':
      return `举报${idText}已驳回`
    case 'REPORT_CLOSE':
      return `举报${idText}已关闭`
    case 'ORDER_EXCEPTION_CLOSE':
      return replaced || `订单${idText}已异常关闭`
    case 'HANDLE_APPEAL':
      return replaced || `申诉${idText}已处理`
    default:
      return replaced || `${target}${idText}已处理`
  }
}

const sourceText = (log) => {
  if (!log?.relatedType || !log?.relatedId) return ''
  const typeText = targetLabelMap[log.relatedType] || log.relatedType
  return `来源：${typeText} #${log.relatedId}`
}

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
  if (log.targetType === 'REPORT' && log.targetId) {
    return `/pages/admin-risk-detail/admin-risk-detail?mode=REPORT&id=${log.targetId}`
  }
  if (log.targetType === 'APPEAL' && log.targetId) {
    return `/pages/admin-risk-detail/admin-risk-detail?mode=APPEAL&id=${log.targetId}`
  }
  if (['STUDENT_VERIFY_APPROVE', 'STUDENT_VERIFY_REJECT'].includes(log.operationType)) {
    return '/pages/admin-students/admin-students'
  }
  const map = {
    USER: '/pages/admin-users/admin-users',
    PRODUCT: '/pages/admin-products/admin-products',
    NOTICE: '/pages/admin-announcements/admin-announcements',
    WORD: '/pages/admin-sensitive/admin-sensitive',
    CATEGORY: '/pages/admin-categories/admin-categories',
    ORDER: '/pages/admin-orders/admin-orders'
  }
  return withSourceQuery(map[log.targetType] || '', log)
}

const withSourceQuery = (url, log) => {
  if (!url || !log?.relatedType || !log?.relatedId) return url
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}relatedType=${log.relatedType}&relatedId=${log.relatedId}`
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

.filter-row {
  width: 100%;
  margin: 4rpx 0 22rpx;
  white-space: nowrap;
}

.filter-inner {
  display: inline-flex;
  gap: 14rpx;
  min-width: 100%;
}

.filter-chip {
  flex-shrink: 0;
  padding: 14rpx 22rpx;
  border-radius: 999rpx;
  background: #fff;
  color: #667085;
  font-size: 26rpx;
}

.filter-chip.active {
  background: #17a84b;
  color: #fff;
  font-weight: 700;
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

.source {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #17a84b;
}

.arrow {
  color: #b2bdca;
  font-size: 42rpx;
  line-height: 42rpx;
}
</style>
