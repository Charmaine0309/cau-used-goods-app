<template>
  <view class="page">
    <view v-if="activeTab === 'data'" class="content">
      <view class="page-title">数据</view>

      <view class="todo-summary">
        <view class="todo-summary-title">核心数据</view>
        <view class="todo-summary-grid">
          <view class="todo-summary-item">
            <view class="todo-summary-value">{{ money(orderOverview.totalCompletedAmount) }}</view>
            <view class="todo-summary-label">成交金额</view>
          </view>
          <view class="todo-summary-item">
            <view class="todo-summary-value">{{ orderOverview.completedOrders || 0 }}</view>
            <view class="todo-summary-label">成交订单</view>
          </view>
        </view>
      </view>

      <view class="section-title">平台概况</view>
      <view class="overview-grid">
        <view class="overview-card">
          <view class="overview-head">
            <text class="overview-title">用户情况</text>
            <text class="overview-tag">认证率 {{ percent(userOverview.verifiedUsers, userOverview.totalUsers) }}</text>
          </view>
          <view class="overview-main">{{ userOverview.totalUsers || 0 }}</view>
          <view class="overview-sub">注册用户</view>
          <view class="overview-row">
            <text>已认证</text>
            <text>{{ userOverview.verifiedUsers || 0 }}</text>
          </view>
          <view class="overview-row">
            <text>审核中</text>
            <text>{{ userOverview.pendingUsers || 0 }}</text>
          </view>
        </view>

        <view class="overview-card">
          <view class="overview-head">
            <text class="overview-title">商品流通</text>
            <text class="overview-tag">在售 {{ percent(productOverview.onSaleProducts, productOverview.totalProducts) }}</text>
          </view>
          <view class="overview-main">{{ productOverview.onSaleProducts || 0 }}</view>
          <view class="overview-sub">在售商品</view>
          <view class="overview-row">
            <text>已下架</text>
            <text>{{ productOverview.offShelfProducts || 0 }}</text>
          </view>
          <view class="overview-row">
            <text>已成交</text>
            <text>{{ productOverview.soldProducts || 0 }}</text>
          </view>
        </view>

        <view class="overview-card">
          <view class="overview-head">
            <text class="overview-title">商品热度</text>
            <text class="overview-tag">浏览 / 收藏</text>
          </view>
          <view class="overview-main">{{ productOverview.totalViews || 0 }}</view>
          <view class="overview-sub">浏览量</view>
          <view class="overview-row">
            <text>收藏数</text>
            <text>{{ productOverview.totalFavorites || 0 }}</text>
          </view>
          <view class="overview-row">
            <text>平均价格</text>
            <text>{{ money(productOverview.averagePrice) }}</text>
          </view>
        </view>

        <view class="overview-card">
          <view class="overview-head">
            <text class="overview-title">风控情况</text>
            <text :class="['overview-tag', reportOverview.pendingReports ? 'danger' : '']">
              {{ reportOverview.pendingReports ? '需处理' : '正常' }}
            </text>
          </view>
          <view :class="['overview-main', riskTodoCount ? 'danger' : '']">
            {{ riskTodoCount }}
          </view>
          <view class="overview-sub">待处理申诉和举报</view>
          <view class="overview-row">
            <text>商品举报</text>
            <text>{{ reportOverview.productReports || 0 }}</text>
          </view>
          <view class="overview-row">
            <text>用户举报</text>
            <text>{{ reportOverview.userReports || 0 }}</text>
          </view>
        </view>
      </view>

      <view class="section-title">近 7 天发布 / 成交趋势</view>
      <view class="trend-card">
        <view class="trend-legend">
          <view class="legend-item"><text class="legend-dot publish"></text>发布</view>
          <view class="legend-item"><text class="legend-dot deal"></text>成交</view>
        </view>
        <view class="trend-bars">
          <view v-for="item in productTrend" :key="item.date" class="trend-item">
            <view class="trend-bar-wrap">
              <view class="trend-bar publish" :style="trendBarStyle(item.count)"></view>
              <view class="trend-bar deal" :style="trendBarStyle(item.completedCount || 0)"></view>
            </view>
            <view class="trend-count">{{ item.count || 0 }}/{{ item.completedCount || 0 }}</view>
            <view class="trend-date">{{ shortDate(item.date) }}</view>
          </view>
        </view>
        <view v-if="!productTrend.length" class="empty">暂无趋势数据</view>
        <view class="trend-note">成交趋势需后端新增按日期统计后接入</view>
      </view>

      <view class="section-title">商品分类统计</view>
      <view class="category-list">
        <view v-for="item in categoryDistribution" :key="item.categoryId" class="category-item">
          <view class="category-head">
            <text class="category-name">{{ item.categoryName }}</text>
            <text class="category-count">{{ item.productCount || 0 }} 件</text>
          </view>
          <view class="category-progress">
            <view class="category-bar" :style="categoryBarStyle(item.productCount)"></view>
            <view class="category-bar sale" :style="categoryOnSaleBarStyle(item)"></view>
          </view>
          <view class="category-meta">
            <text>在售 {{ item.onSaleCount || 0 }}，占比 {{ percent(item.onSaleCount, item.productCount) }}</text>
            <text>均价 {{ money(item.averagePrice) }}</text>
          </view>
        </view>
        <view v-if="!categoryDistribution.length" class="empty">暂无分类统计</view>
      </view>

      <view class="section-title">运营提醒</view>
      <view class="panel">
        <view class="row">
          <text>待确认订单</text>
          <text>{{ orderOverview.pendingConfirmOrders || 0 }}</text>
        </view>
        <view class="row">
          <text>待面交订单</text>
          <text>{{ orderOverview.waitMeetOrders || 0 }}</text>
        </view>
        <view class="row">
          <text>已下架商品</text>
          <text>{{ productOverview.offShelfProducts || 0 }}</text>
        </view>
        <view class="row">
          <text>异常关闭订单</text>
          <text>{{ orderOverview.exceptionClosedOrders || 0 }}</text>
        </view>
      </view>
    </view>

    <view v-else-if="activeTab === 'review'" class="content">
      <view class="page-title">审核</view>
      <view class="page-subtitle">点击待办卡片进入对应处理页面</view>

      <view class="todo-panel">
        <view class="todo-card urgent" @click="goPage('/pages/admin-students/admin-students')">
          <view class="todo-number">{{ userOverview.pendingUsers || 0 }}</view>
          <view class="todo-label">待认证审核</view>
        </view>
        <view class="todo-card" @click="goPage('/pages/admin-reports/admin-reports')">
          <view class="todo-number">{{ riskTodoCount }}</view>
          <view class="todo-label">待处理申诉和举报</view>
        </view>
      </view>

      <view class="section-title">管理</view>
      <view class="manage-grid">
        <view class="manage-card" @click="goPage('/pages/admin-products/admin-products')">
          <view class="manage-title">商品管理</view>
          <view class="manage-desc">上架 / 下架</view>
        </view>
        <view class="manage-card" @click="goPage('/pages/admin-users/admin-users')">
          <view class="manage-title">用户管理</view>
          <view class="manage-desc">用户与角色</view>
        </view>
        <view class="manage-card" @click="goPage('/pages/admin-sensitive/admin-sensitive')">
          <view class="manage-title">敏感词管理</view>
          <view class="manage-desc">审核词库</view>
        </view>
        <view class="manage-card" @click="goPage('/pages/admin-categories/admin-categories')">
          <view class="manage-title">标签管理</view>
          <view class="manage-desc">商品分类标签</view>
        </view>
        <view class="manage-card" @click="goPage('/pages/admin-announcements/admin-announcements')">
          <view class="manage-title">公告管理</view>
          <view class="manage-desc">发布 / 下线</view>
        </view>
        <view class="manage-card" @click="goPage('/pages/admin-stats/admin-stats')">
          <view class="manage-title">申诉统计</view>
          <view class="manage-desc">申诉数据</view>
        </view>
      </view>

      <view class="section-title">说明</view>
      <view class="hint-panel">
        <view class="hint-row">
          <text class="hint-dot"></text>
          <text>待处理事项优先处理，处理后数量会自动更新。</text>
        </view>
        <view class="hint-row">
          <text class="hint-dot"></text>
          <text>商品、用户、公告和敏感词属于日常管理入口。</text>
        </view>
      </view>
    </view>

    <view v-else class="content">
      <view class="page-title">我的</view>
      <view class="profile-card">
        <image v-if="adminAvatarUrl" class="avatar" :src="adminAvatarUrl" mode="aspectFill" />
        <view v-else class="avatar placeholder">管</view>
        <view class="profile-main">
          <view class="nickname">{{ currentUser.nickname || '管理员' }}</view>
          <view class="role">管理员</view>
        </view>
      </view>

      <view class="menu-list">
        <view class="menu-item" @click="goPage('/pages/profile-edit/profile-edit')">
          <view class="menu-title">个人资料编辑</view>
          <text class="arrow">›</text>
        </view>
        <view class="menu-item" @click="goPage('/pages/admin-logs/admin-logs')">
          <view class="menu-title">管理员日志</view>
          <text class="arrow">›</text>
        </view>
      </view>

      <button class="logout-button" @click="logout">退出登录</button>
    </view>

    <view class="tabbar">
      <view :class="['tab-item', activeTab === 'data' ? 'active' : '']" @click="switchTab('data')">数据</view>
      <view :class="['tab-item', activeTab === 'review' ? 'active' : '']" @click="switchTab('review')">审核</view>
      <view :class="['tab-item', activeTab === 'mine' ? 'active' : '']" @click="switchTab('mine')">我的</view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  getCategoryDistribution,
  getAppealOverview,
  getOrderOverview,
  getProductOverview,
  getProductTrend,
  getReportOverview,
  getUserOverview
} from '../../api/admin'
import { getCurrentUser } from '../../api/auth'
import { clearAuth, getUser, setUser } from '../../utils/auth'
import { BASE_URL } from '../../utils/request'

const activeTab = ref('data')
const currentUser = ref(getUser() || {})
const userOverview = ref({})
const productOverview = ref({})
const orderOverview = ref({})
const reportOverview = ref({})
const appealOverview = ref({})
const categoryDistribution = ref([])
const productTrend = ref([])

const riskTodoCount = computed(() => {
  return Number(reportOverview.value.pendingReports || 0) + Number(appealOverview.value.pendingAppeals || 0)
})

const adminAvatarUrl = computed(() => {
  const url = currentUser.value?.avatarUrl || currentUser.value?.avatar_url || ''
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (url.startsWith('/uploads/')) return BASE_URL + url
  return url
})

const loadAdminData = async () => {
  try {
    const [me, users, productStats, orders, reportStats, appealStats, categories, trend] = await Promise.all([
      getCurrentUser(),
      getUserOverview(),
      getProductOverview(),
      getOrderOverview(),
      getReportOverview(),
      getAppealOverview(),
      getCategoryDistribution(),
      getProductTrend(7)
    ])
    currentUser.value = me || {}
    setUser(me || {})
    userOverview.value = users || {}
    productOverview.value = productStats || {}
    orderOverview.value = orders || {}
    reportOverview.value = reportStats || {}
    appealOverview.value = appealStats || {}
    categoryDistribution.value = categories || []
    productTrend.value = trend?.list || trend || []
  } catch (error) {
    uni.showToast({ title: error.message || '后台数据加载失败', icon: 'none' })
  }
}

onShow(loadAdminData)

const switchTab = (tab) => {
  activeTab.value = tab
}

const goPage = (url) => {
  uni.navigateTo({ url })
}

const logout = () => {
  clearAuth()
  uni.reLaunch({ url: '/pages/login/login' })
}

const money = (value) => {
  return `￥${Number(value || 0).toFixed(2)}`
}

const percent = (value, total) => {
  const num = Number(value || 0)
  const den = Number(total || 0)
  if (!den) return '0%'
  return `${Math.round((num / den) * 100)}%`
}

const trendBarStyle = (count) => {
  const max = Math.max(...productTrend.value.flatMap((item) => [
    Number(item.count || 0),
    Number(item.completedCount || 0)
  ]), 1)
  const height = Math.max(12, Math.round((Number(count || 0) / max) * 120))
  return `height: ${height}rpx;`
}

const shortDate = (date) => {
  if (!date) return ''
  const value = String(date)
  const datePart = value.includes('T') ? value.split('T')[0] : value
  const parts = datePart.split('-')
  if (parts.length < 3) return value
  return `${parts[1]}.${parts[2]}`
}

const categoryBarStyle = (count) => {
  const max = Math.max(...categoryDistribution.value.map((item) => Number(item.productCount || 0)), 1)
  const width = Math.max(8, Math.round((Number(count || 0) / max) * 100))
  return `width: ${width}%;`
}

const categoryOnSaleBarStyle = (item) => {
  const total = Number(item?.productCount || 0)
  const onSale = Number(item?.onSaleCount || 0)
  if (!total) return 'width: 0%;'
  return `width: ${Math.round((onSale / total) * 100)}%;`
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding-bottom: 140rpx;
  background: #f5f6f8;
  box-sizing: border-box;
}

.content {
  padding: 28rpx 24rpx;
}

.page-title {
  margin: 18rpx 0 24rpx;
  font-size: 38rpx;
  font-weight: 700;
  color: #1f2933;
}

.page-subtitle {
  margin: -10rpx 0 22rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.section-title {
  margin: 32rpx 0 18rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2933;
}

.todo-summary {
  padding: 30rpx;
  border-radius: 18rpx;
  background: #14213d;
  box-sizing: border-box;
}

.todo-summary-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #fff;
}

.todo-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
  margin-top: 22rpx;
}

.todo-summary-item {
  min-height: 150rpx;
  padding: 24rpx;
  border-radius: 16rpx;
  background: rgba(255, 255, 255, 0.1);
  box-sizing: border-box;
}

.todo-summary-value {
  font-size: 58rpx;
  line-height: 64rpx;
  font-weight: 700;
  color: #dcfce7;
}

.todo-summary-value.warning {
  color: #fde68a;
}

.todo-summary-value.danger {
  color: #fca5a5;
}

.todo-summary-label {
  margin-top: 14rpx;
  font-size: 26rpx;
  color: #d0d5dd;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
}

.overview-card {
  min-height: 244rpx;
  padding: 24rpx;
  border-radius: 16rpx;
  background: #fff;
  box-sizing: border-box;
}

.overview-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10rpx;
}

.overview-title {
  font-size: 26rpx;
  font-weight: 700;
  color: #1f2933;
}

.overview-tag {
  max-width: 150rpx;
  padding: 4rpx 10rpx;
  border-radius: 999rpx;
  background: #dcfce7;
  color: #16a34a;
  font-size: 20rpx;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.overview-tag.danger {
  background: #fee2e2;
  color: #ef4444;
}

.overview-main {
  margin-top: 20rpx;
  font-size: 52rpx;
  line-height: 58rpx;
  font-weight: 700;
  color: #1f2933;
}

.overview-main.danger {
  color: #ef4444;
}

.overview-sub {
  margin-top: 8rpx;
  margin-bottom: 14rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.overview-row {
  display: flex;
  justify-content: space-between;
  padding-top: 10rpx;
  font-size: 24rpx;
  color: #667085;
}

.quick-list {
  border-radius: 16rpx;
  background: #fff;
  overflow: hidden;
}

.quick-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 116rpx;
  padding: 24rpx 28rpx;
  border-bottom: 1rpx solid #eef0f3;
  box-sizing: border-box;
}

.quick-item:last-child {
  border-bottom: 0;
}

.quick-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #1f2933;
}

.quick-desc {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.quick-count {
  min-width: 44rpx;
  color: #17a84b;
  font-size: 34rpx;
  font-weight: 700;
  text-align: right;
}

.quick-count.danger {
  color: #ef4444;
}

.trend-card {
  padding: 28rpx 22rpx;
  border-radius: 16rpx;
  background: #fff;
}

.trend-legend {
  display: flex;
  justify-content: flex-end;
  gap: 24rpx;
  margin-bottom: 18rpx;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
  font-size: 22rpx;
  color: #667085;
}

.legend-dot {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
}

.legend-dot.publish {
  background: #17a84b;
}

.legend-dot.deal {
  background: #3b82f6;
}

.trend-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  min-height: 220rpx;
}

.trend-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 0;
}

.trend-bar-wrap {
  display: flex;
  align-items: flex-end;
  gap: 6rpx;
  height: 132rpx;
}

.trend-bar {
  width: 22rpx;
  min-height: 12rpx;
  border-radius: 999rpx 999rpx 8rpx 8rpx;
}

.trend-bar.publish {
  background: linear-gradient(180deg, #22c55e 0%, #16a34a 100%);
}

.trend-bar.deal {
  background: linear-gradient(180deg, #60a5fa 0%, #2563eb 100%);
}

.trend-count {
  margin-top: 12rpx;
  font-size: 20rpx;
  color: #1f2933;
  font-weight: 700;
}

.trend-date {
  margin-top: 6rpx;
  font-size: 20rpx;
  color: #98a2b3;
}

.trend-note {
  margin-top: 18rpx;
  padding-top: 18rpx;
  border-top: 1rpx solid #eef0f3;
  font-size: 22rpx;
  color: #98a2b3;
}

.category-list {
  border-radius: 16rpx;
  background: #fff;
  overflow: hidden;
}

.category-item {
  padding: 24rpx 28rpx;
  border-bottom: 1rpx solid #eef0f3;
}

.category-item:last-child {
  border-bottom: 0;
}

.category-head,
.category-meta {
  display: flex;
  justify-content: space-between;
  gap: 20rpx;
}

.category-name {
  font-size: 28rpx;
  font-weight: 700;
  color: #1f2933;
}

.category-count {
  font-size: 26rpx;
  color: #17a84b;
}

.category-progress {
  position: relative;
  height: 12rpx;
  margin: 18rpx 0 14rpx;
  border-radius: 999rpx;
  background: #eef2f6;
  overflow: hidden;
}

.category-bar {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  border-radius: 999rpx;
  background: #17a84b;
}

.category-bar.sale {
  background: #facc15;
}

.category-meta {
  font-size: 22rpx;
  color: #8a96a8;
}

.todo-panel {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
}

.todo-card {
  min-height: 150rpx;
  padding: 28rpx;
  border-radius: 16rpx;
  background: #fff;
  border: 2rpx solid transparent;
  box-sizing: border-box;
}

.todo-card.urgent {
  border-color: #22c55e;
}

.todo-number {
  font-size: 50rpx;
  line-height: 56rpx;
  font-weight: 700;
  color: #17a84b;
}

.todo-label {
  margin-top: 14rpx;
  font-size: 26rpx;
  color: #667085;
}

.panel,
.menu-list,
.profile-card {
  border-radius: 16rpx;
  background: #fff;
  overflow: hidden;
}

.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx 28rpx;
  border-bottom: 1rpx solid #eef0f3;
  font-size: 28rpx;
  color: #475467;
}

.row text:first-child {
  flex: 1;
}

.row text:last-child {
  min-width: 52rpx;
  height: 44rpx;
  padding: 0 16rpx;
  border-radius: 999rpx;
  background: #f0fdf4;
  color: #16a34a;
  font-weight: 700;
  line-height: 44rpx;
  text-align: center;
}

.row:last-child {
  border-bottom: 0;
}

.empty {
  padding: 40rpx 28rpx;
  font-size: 28rpx;
  color: #98a2b3;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 118rpx;
  padding: 24rpx 28rpx;
  border-bottom: 1rpx solid #eef0f3;
  box-sizing: border-box;
}

.menu-item:last-child {
  border-bottom: 0;
}

.menu-item.primary {
  min-height: 132rpx;
}

.menu-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #1f2933;
}

.menu-desc {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.menu-side {
  display: flex;
  align-items: center;
  gap: 14rpx;
}

.badge {
  min-width: 38rpx;
  height: 38rpx;
  padding: 0 12rpx;
  border-radius: 999rpx;
  background: #fee2e2;
  color: #ef4444;
  font-size: 24rpx;
  line-height: 38rpx;
  text-align: center;
}

.manage-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
}

.manage-card {
  min-height: 138rpx;
  padding: 26rpx;
  border-radius: 16rpx;
  background: #fff;
  box-sizing: border-box;
}

.manage-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #1f2933;
}

.manage-desc {
  margin-top: 14rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.hint-panel {
  padding: 24rpx 28rpx;
  border-radius: 16rpx;
  background: #fff;
}

.hint-row {
  display: flex;
  align-items: flex-start;
  gap: 14rpx;
  margin-bottom: 16rpx;
  font-size: 24rpx;
  line-height: 36rpx;
  color: #667085;
}

.hint-row:last-child {
  margin-bottom: 0;
}

.hint-dot {
  width: 12rpx;
  height: 12rpx;
  margin-top: 12rpx;
  border-radius: 50%;
  background: #17a84b;
  flex-shrink: 0;
}

.arrow {
  font-size: 42rpx;
  color: #b2bdca;
}

.profile-card {
  display: flex;
  align-items: center;
  padding: 30rpx;
  margin-bottom: 24rpx;
}

.avatar {
  width: 108rpx;
  height: 108rpx;
  border-radius: 54rpx;
  margin-right: 24rpx;
  background: #e7edf3;
}

.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #17a84b;
  font-size: 36rpx;
  font-weight: 700;
}

.profile-main {
  flex: 1;
  min-width: 0;
}

.nickname {
  font-size: 34rpx;
  font-weight: 700;
  color: #1f2933;
}

.role {
  margin-top: 10rpx;
  font-size: 26rpx;
  color: #667085;
}

.logout-button {
  margin-top: 28rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 12rpx;
  background: #fff;
  color: #ef4444;
  font-size: 30rpx;
}

.tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  display: flex;
  height: 112rpx;
  padding-bottom: env(safe-area-inset-bottom);
  background: #fff;
  border-top: 1rpx solid #eef0f3;
}

.tab-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #667085;
}

.tab-item.active {
  color: #17a84b;
  font-weight: 700;
}
</style>
