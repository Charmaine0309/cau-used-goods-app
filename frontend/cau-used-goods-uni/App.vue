<script>
export default {
  onLaunch() {
    console.log('App Launch')
  },
  onShow() {
    console.log('App Show')
  },
  onHide() {
    console.log('App Hide')
  }
}
</script>

<style lang="scss">
@import './styles/trade.scss';

page {
  background: #f6f8f5;
  color: #24302b;
  font-size: 28rpx;
}

button::after {
  border: 0;
}

.empty-state {
  padding: 96rpx 24rpx;
  color: #89938f;
  text-align: center;
}
</style>
