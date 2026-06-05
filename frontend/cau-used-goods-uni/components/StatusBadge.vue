<template>
  <text class="badge" :class="`badge-${tone}`">{{ label }}</text>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  tone: { type: String, default: 'muted' }
})
</script>

<style scoped lang="scss">
.badge {
  display: inline-flex;
  padding: 6rpx 14rpx;
  border-radius: 20rpx;
  font-size: 22rpx;
}

.badge-warning { color: #996515; background: #fff3d6; }
.badge-primary { color: #2f6b4f; background: #e7f4ed; }
.badge-success { color: #267247; background: #dff4e8; }
.badge-danger { color: #ba3d39; background: #fde8e7; }
.badge-muted { color: #738077; background: #eef1ef; }
</style>
