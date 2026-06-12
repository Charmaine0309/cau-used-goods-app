<template>
  <view class="page">
    <view v-if="!canPublish">
      <view class="locked-card">
        <view class="locked-title">&#20808;&#23436;&#25104;&#35748;&#35777;&#65292;&#20877;&#21457;&#24067;&#38386;&#32622;</view>
        <view class="locked-desc">&#20320;&#30340;&#23398;&#29983;&#35748;&#35777;&#23578;&#26410;&#36890;&#36807;&#65292;&#21457;&#24067;&#34920;&#21333;&#24050;&#26242;&#26102;&#20851;&#38381;&#12290;&#35748;&#35777;&#36890;&#36807;&#21518;&#21363;&#21487;&#21457;&#24067;&#21830;&#21697;&#12290;</view>
        <button class="locked-button" @click="goStudentAuth">&#21435;&#23398;&#29983;&#35748;&#35777;</button>
      </view>
    </view>

    <view v-else>
      <view class="hero">
        <text class="eyebrow">CAU CAMPUS MARKET</text>
        <text class="title">&#21457;&#24067;&#20320;&#30340;&#38386;&#32622;&#22909;&#29289;</text>
        <text class="hero-copy">&#30495;&#23454;&#25551;&#36848;&#29289;&#21697;&#24773;&#20917;&#65292;&#26356;&#23481;&#26131;&#36935;&#21040;&#21512;&#36866;&#30340;&#26032;&#20027;&#20154;</text>
      </view>
      <view class="card">
        <view class="field">
          <view class="field-head">
            <text>&#21830;&#21697;&#22270;&#29255;</text>
            <text class="muted">{{ form.images.length }}/9</text>
          </view>
          <view class="image-grid">
            <view v-for="(image,index) in form.images" :key="image + index" class="image-item">
              <image :src="image" mode="aspectFill" />
              <text class="remove" @click="removeImage(index)">x</text>
            </view>
            <view v-if="form.images.length < 9" class="upload" @click="chooseImages">
              +
              <text>&#28155;&#21152;&#22270;&#29255;</text>
            </view>
          </view>
          <text class="hint">&#26368;&#22810;&#19978;&#20256; 9 &#24352;&#22270;&#29255;&#65292;&#21333;&#24352;&#19981;&#36229;&#36807; 5 MB&#65292;&#31532;&#19968;&#24352;&#20316;&#20026;&#23553;&#38754;&#12290;</text>
        </view>

        <view class="field">
          <view class="field-head">
            <text>&#21830;&#21697;&#26631;&#39064;</text>
            <text class="ai" @click="optimizeTitle">AI &#20248;&#21270;&#26631;&#39064;</text>
          </view>
          <input v-model.trim="form.title" maxlength="100" placeholder="&#20363;&#22914;&#65306;&#20061;&#25104;&#26032;&#23567;&#31859;&#21488;&#28783;&#65292;&#23487;&#33293;&#33258;&#29992;" />
        </view>

        <view class="field inline">
          <text>&#20998;&#31867;</text>
          <picker :range="categories" range-key="name" @change="changeCategory">
            <view class="picker">{{ categoryName || '\u8bf7\u9009\u62e9\u5206\u7c7b' }} &#8250;</view>
          </picker>
        </view>

        <view class="field inline">
          <text>&#25104;&#33394;</text>
          <picker :range="conditions" @change="changeCondition">
            <view class="picker">{{ form.conditionLevel || '\u8bf7\u9009\u62e9\u6210\u8272' }} &#8250;</view>
          </picker>
        </view>

        <view class="price-row">
          <view class="field">
            <text>&#21407;&#20215;</text>
            <input v-model="form.originalPrice" type="digit" placeholder="&#36873;&#22635;" />
          </view>
          <view class="field">
            <text>&#21806;&#20215;</text>
            <input v-model="form.price" type="digit" placeholder="&#24517;&#22635;" />
          </view>
        </view>

        <view class="field">
          <view class="field-head">
            <text>&#21830;&#21697;&#25551;&#36848;</text>
            <text class="ai" @click="generateDescription">AI &#29983;&#25104;&#25551;&#36848;</text>
          </view>
          <textarea v-model.trim="form.description" maxlength="1000" placeholder="&#35828;&#26126;&#20351;&#29992;&#24773;&#20917;&#12289;&#22806;&#35266;&#29808;&#30133;&#12289;&#37197;&#20214;&#31561;&#20449;&#24687;" />
        </view>

        <view class="field">
          <text>&#24314;&#35758;&#38754;&#20132;&#22320;&#28857;</text>
          <input v-model.trim="form.meetLocation" maxlength="100" placeholder="&#20363;&#22914;&#65306;&#19996;&#21306;&#22270;&#20070;&#39302;&#38376;&#21475;" />
        </view>
      </view>
      <button class="submit" :loading="submitting" @click="submit">&#30830;&#35748;&#21457;&#24067;</button>
    </view>
  </view>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { createProduct, generateProductDescription, listCategories, optimizeProductTitle, uploadProductImage } from '../../api/product'
import { getCurrentUser } from '../../api/auth'
import { getToken, setUser } from '../../utils/auth'

const MAX_SIZE = 5 * 1024 * 1024
const conditions = ['\u5168\u65b0', '\u4e5d\u6210\u65b0', '\u516b\u6210\u65b0', '\u4e03\u6210\u65b0', '\u6709\u660e\u663e\u4f7f\u7528\u75d5\u8ff9']
const categories = ref([])
const submitting = ref(false)
const canPublish = ref(false)
const form = reactive({ images: [], title: '', categoryId: '', conditionLevel: '', originalPrice: '', price: '', description: '', meetLocation: '' })

const categoryName = computed(() => categories.value.find((item) => item.id === Number(form.categoryId))?.name || '')
const toast = (title) => uni.showToast({ title, icon: 'none' })

const isVerified = (user) => {
  const status = user?.authStatus || user?.auth_status || ''
  return status === 'VERIFIED'
}

const loadPublishState = async () => {
  if (!getToken()) {
    canPublish.value = false
    uni.showToast({ title: '\u8bf7\u5148\u767b\u5f55', icon: 'none' })
    return
  }
  try {
    const current = await getCurrentUser()
    setUser(current)
    canPublish.value = isVerified(current)
    if (!canPublish.value) {
      uni.showToast({ title: '\u672a\u5b8c\u6210\u5b66\u751f\u8ba4\u8bc1', icon: 'none' })
      return
    }
    categories.value = (await listCategories()).filter((item) => Number(item.id) !== 0)
  } catch (error) {
    canPublish.value = false
    toast(error.message || '\u52a0\u8f7d\u5931\u8d25')
  }
}

onShow(loadPublishState)

const goStudentAuth = () => uni.navigateTo({ url: '/pages/student-auth/student-auth' })
const changeCategory = ({ detail }) => { form.categoryId = categories.value[detail.value].id }
const changeCondition = ({ detail }) => { form.conditionLevel = conditions[detail.value] }
const removeImage = (index) => form.images.splice(index, 1)

const chooseImages = async () => {
  if (!canPublish.value) return toast('\u672a\u5b8c\u6210\u5b66\u751f\u8ba4\u8bc1')
  try {
    const result = await uni.chooseImage({ count: 9 - form.images.length, sizeType: ['compressed'] })
    const files = result.tempFiles.filter((file) => {
      if (file.size > MAX_SIZE) {
        toast('\u5355\u5f20\u56fe\u7247\u4e0d\u80fd\u8d85\u8fc7 5 MB')
        return false
      }
      return true
    })
    if (!files.length) return
    uni.showLoading({ title: '\u4e0a\u4f20\u56fe\u7247\u4e2d' })
    for (const file of files) {
      const uploaded = await uploadProductImage(file.path)
      form.images.push(uploaded.imageUrl)
    }
  } catch (error) {
    if (!error?.errMsg?.includes('cancel')) toast(error.message || '\u56fe\u7247\u4e0a\u4f20\u5931\u8d25')
  } finally {
    uni.hideLoading()
  }
}

const optimizeTitle = async () => {
  if (!canPublish.value) return toast('\u672a\u5b8c\u6210\u5b66\u751f\u8ba4\u8bc1')
  if (!form.title) return toast('\u8bf7\u5148\u586b\u5199\u4e00\u4e2a\u57fa\u7840\u6807\u9898')
  try {
    uni.showLoading({ title: 'AI \u6b63\u5728\u4f18\u5316' })
    const { titles } = await optimizeProductTitle({ title: form.title, categoryName: categoryName.value, conditionLevel: form.conditionLevel })
    uni.showActionSheet({ itemList: titles, success: ({ tapIndex }) => { form.title = titles[tapIndex] } })
  } catch (error) {
    toast('AI \u670d\u52a1\u6682\u65f6\u4e0d\u53ef\u7528')
  } finally {
    uni.hideLoading()
  }
}

const generateDescription = async () => {
  if (!canPublish.value) return toast('\u672a\u5b8c\u6210\u5b66\u751f\u8ba4\u8bc1')
  if (!form.title) return toast('\u8bf7\u5148\u586b\u5199\u6807\u9898')
  try {
    uni.showLoading({ title: 'AI \u6b63\u5728\u751f\u6210' })
    const result = await generateProductDescription({ title: form.title, categoryName: categoryName.value, conditionLevel: form.conditionLevel, meetLocation: form.meetLocation })
    form.description = result.description
  } catch (error) {
    toast('AI \u670d\u52a1\u6682\u65f6\u4e0d\u53ef\u7528')
  } finally {
    uni.hideLoading()
  }
}

const validate = () => {
  if (!canPublish.value) return '\u672a\u5b8c\u6210\u5b66\u751f\u8ba4\u8bc1'
  if (!form.images.length) return '\u8bf7\u81f3\u5c11\u4e0a\u4f20\u4e00\u5f20\u5546\u54c1\u56fe\u7247'
  if (!form.title) return '\u8bf7\u586b\u5199\u5546\u54c1\u6807\u9898'
  if (!form.categoryId) return '\u8bf7\u9009\u62e9\u5546\u54c1\u5206\u7c7b'
  if (!form.conditionLevel) return '\u8bf7\u9009\u62e9\u5546\u54c1\u6210\u8272'
  if (!form.price || Number(form.price) <= 0) return '\u8bf7\u586b\u5199\u6b63\u786e\u7684\u5546\u54c1\u552e\u4ef7'
  if (!form.description) return '\u8bf7\u586b\u5199\u5546\u54c1\u63cf\u8ff0'
  return ''
}

const submit = async () => {
  const message = validate()
  if (message) return toast(message)
  submitting.value = true
  try {
    await createProduct({ ...form, price: Number(form.price), originalPrice: Number(form.originalPrice || 0) })
    uni.setStorageSync('PRODUCT_LIST_DIRTY', true)
    uni.showToast({ title: '\u53d1\u5e03\u6210\u529f', icon: 'success' })
    setTimeout(() => uni.switchTab({ url: '/pages/home/home' }), 600)
  } catch (error) {
    toast(error.message || '\u53d1\u5e03\u5931\u8d25')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding-bottom: 36rpx;
  background: #f5f8f6;
  box-sizing: border-box;
}

.hero {
  margin: 24rpx 28rpx 0;
  padding: 34rpx 30rpx;
  border-radius: 28rpx;
  background: linear-gradient(145deg, #1f6a49, #328660);
  color: #fff;
}

.locked-hero {
  padding-bottom: 34rpx;
}

.eyebrow,
.title,
.headline,
.hero-copy,
.muted,
.hint {
  display: block;
}

.eyebrow {
  color: rgba(255,255,255,.7);
  font-size: 20rpx;
  letter-spacing: 3rpx;
}

.title,
.headline {
  margin-top: 14rpx;
  font-size: 38rpx;
  font-weight: 700;
}

.hero-copy {
  margin-top: 18rpx;
  color: rgba(255,255,255,.78);
  font-size: 24rpx;
  line-height: 1.6;
}

.locked-card {
  position: relative;
  margin: 32rpx 28rpx 0;
  padding: 42rpx 34rpx 36rpx;
  border-radius: 28rpx;
  background: #fff;
  box-shadow: 0 18rpx 44rpx rgba(31, 106, 73, .12);
}

.locked-title {
  color: #1f2933;
  font-size: 34rpx;
  line-height: 44rpx;
  font-weight: 700;
}

.locked-desc {
  margin-top: 16rpx;
  color: #667085;
  font-size: 26rpx;
  line-height: 1.65;
}

.locked-button {
  margin-top: 34rpx;
  height: 84rpx;
  line-height: 84rpx;
  border-radius: 999rpx;
  background: #23734f;
  color: #fff;
  font-size: 28rpx;
}

.card {
  margin: 24rpx 28rpx 0;
  padding: 4rpx 24rpx;
  border-radius: 22rpx;
  background: #fff;
}

.field {
  padding: 24rpx 0;
  border-bottom: 2rpx solid #edf2ef;
}

.field:last-child {
  border-bottom: 0;
}

.field-head,
.inline {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.field text,
.field-head text:first-child,
.inline text {
  color: #26342f;
  font-size: 28rpx;
  font-weight: 700;
}

.ai,
.picker {
  color: #23734f;
  font-size: 25rpx;
  font-weight: 600;
}

.muted,
.hint {
  margin-top: 10rpx;
  color: #929c98;
  font-size: 22rpx;
  line-height: 1.55;
}

input {
  height: 70rpx;
  margin-top: 10rpx;
  color: #1f2933;
  font-size: 27rpx;
}

textarea {
  width: 100%;
  height: 190rpx;
  margin-top: 16rpx;
  padding: 18rpx;
  border-radius: 18rpx;
  background: #f6f8f5;
  color: #1f2933;
  font-size: 26rpx;
  box-sizing: border-box;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14rpx;
  margin-top: 18rpx;
}

.image-item,
.upload {
  position: relative;
  height: 170rpx;
  overflow: hidden;
  border-radius: 18rpx;
}

.image-item image {
  width: 100%;
  height: 100%;
}

.remove {
  position: absolute;
  top: 8rpx;
  right: 8rpx;
  width: 34rpx;
  height: 34rpx;
  border-radius: 50%;
  background: rgba(0,0,0,.58);
  color: #fff;
  text-align: center;
  line-height: 32rpx;
}

.upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 2rpx dashed #aac4b8;
  background: #f3faf6;
  color: #23734f;
  font-size: 42rpx;
}

.upload text {
  margin-top: 6rpx;
  font-size: 22rpx;
}

.price-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24rpx;
}

.submit {
  margin: 28rpx 28rpx 0;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 999rpx;
  background: #23734f;
  color: #fff;
  font-size: 30rpx;
}
</style>
