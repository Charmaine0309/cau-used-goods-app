<template>
  <view class="page">
    <view v-if="reports.length" class="list">
      <view v-for="report in reports" :key="report.id" class="card report">
        <view class="dot" :class="{ done: report.status === 'RESOLVED' }" />
        <view class="report-body">
          <view class="report-head">
            <text class="report-title">{{ report.reasonLabel || report.reason }}</text>
            <StatusBadge :label="status(report.status).label" :tone="status(report.status).tone" />
          </view>
          <text class="report-meta">{{ report.targetTypeLabel || report.targetType }} · {{ report.targetId }} · {{ report.createdAt }}</text>
          <text class="report-detail">{{ report.detail }}</text>
          <view v-if="report.result" class="result">处理结果：{{ report.result }}</view>
        </view>
      </view>
    </view>
    <EmptyState v-else title="暂无举报记录" detail="举报处理进度和结果会显示在这里" />
  </view>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import { tradeService } from '../../services/trade'
import { REPORT_STATUS } from '../../utils/constants'
import { showError } from '../../utils/navigation'

const reports = ref([])
onShow(async () => {
  try {
    reports.value = await tradeService.getReports()
  } catch (error) {
    showError(error)
  }
})

function status(value) {
  return REPORT_STATUS[value] || { label: value, tone: 'muted' }
}
</script>

<style scoped lang="scss">
.page { min-height: 100vh; padding: 24rpx; background: #f6f7f9; }
.list { display: flex; flex-direction: column; gap: 18rpx; }
.report { display: flex; gap: 14rpx; margin-bottom: 0; padding: 24rpx; border-radius: 18rpx; background: #fff; }
.dot { width: 16rpx; height: 16rpx; margin-top: 10rpx; flex: 0 0 16rpx; border-radius: 50%; background: #f2a23a; }
.dot.done { background: #ccd4d0; }
.report-body { display: flex; flex: 1; min-width: 0; flex-direction: column; gap: 12rpx; }
.report-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 14rpx; }
.report-title { flex: 1; min-width: 0; color: #243129; font-size: 29rpx; font-weight: 700; line-height: 1.4; }
.report-meta { color: #98a39d; font-size: 22rpx; line-height: 1.5; }
.report-detail { color: #59675f; font-size: 25rpx; line-height: 1.6; word-break: break-all; }
.result { padding: 16rpx; border-radius: 12rpx; color: #2f6b4f; background: #edf6f1; font-size: 24rpx; line-height: 1.5; }
</style>
