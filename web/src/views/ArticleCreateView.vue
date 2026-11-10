<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'

const router = useRouter()
const title = ref('')
const content = ref('')
const handlePublish = async () => {
    if (!title.value || !content.value) {
        alert('标题和内容不得为空')
        return
    }
    try {
    await request.post('/v1/articles',{
        title : title.value,
        content: content.value
    })
    alert('发布成功！')
    router.push('/articles') //发布成功后跳回列表页
} catch (err){
    alert('发布失败：' + (err.response?.data?.error || '网络错误'))
}
}

</script>
<template>
    <div class="create-box">
        <h2>发布新文章</h2>
        <input v-model="title" type="text" placeholder="请输入文章标题" />
        <textarea v-model="content" rows="6" placeholder="请输入文章内容"></textarea>
        <button @click="handlePublish">发布文章</button>
        <p class="back-text" @click="router.push('/articles')">返回列表</p>
    </div>
</template>

<style scoped>
.create-box {
    width: 500px;
    margin: 50px auto;
    padding: 20px;
    border: 1px solid #ccc;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    gap: 15px;
}
.create-box h2 {
    text-align: center;
    margin: 0;
}
.create-box input, .create-box textarea {
    padding: 10px;
    font-size: 16px;
    border: 1px solid #ddd;
    border-radius: 4px;
}
.create-box button {
    padding: 10px;
    background-color: #42b883;
    color: white;
    border: none;
    cursor: pointer;
    border-radius: 4px;
    font-size: 16px;
}
.back-text {
    text-align: center;
    color: #666;
    cursor: pointer;
    font-size: 14px;
}
</style>