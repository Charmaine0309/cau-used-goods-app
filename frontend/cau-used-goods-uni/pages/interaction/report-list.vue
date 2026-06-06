<template>
  <view class="page">
    <view v-if="reports.length">
      <view v-for="report in reports" :key="report.id" class="card report">
        <view class="report-head">
          <text class="report-title">{{ report.reasonLabel || report.reason }}</text>
          <StatusBadge :label="status(report.status).label" :tone="status(report.status).tone" />
        </view>
        <text class="report-meta">{{ report.targetTypeLabel || report.targetType }} · {{ report.targetId }} · {{ report.createdAt }}</text>
        <text class="report-detail">{{ report.detail }}</text>
        <view v-if="report.result" class="result">处理结果：{{ report.result }}</view>
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
.report { display: flex; flex-direction: column; gap: 12rpx; }
.report-head { display: flex; align-items: center; justify-content: space-between; }
.report-title { font-size: 29rpx; font-weight: 700; }
.report-meta { color: #98a39d; font-size: 22rpx; }
.report-detail { color: #59675f; font-size: 25rpx; line-height: 1.6; }
.result { padding: 16rpx; border-radius: 12rpx; color: #2f6b4f; background: #edf6f1; font-size: 24rpx; line-height: 1.5; }
</style>
