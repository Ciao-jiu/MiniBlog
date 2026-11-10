<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '@/utils/request'
const route = useRoute()
const router = useRouter()

//获取路由参数中的文章 ID
const title = ref('')
const articleId = route.params.id
const content = ref('')
const isLoading = ref(true)

onMounted(async () => {
    try {
        const res = await request.get(`/v1/articles/${articleId}`)

        title.value = res.Title || res.title
        content.value = res.Content || res.content
    } catch (err) {
        alert('获取文章失败:' + (err.response?.data?.error || '网络错误'))
        // 失败了就退回列表页
        router.push('/articles')
    } finally {
        isLoading.value = false
    }
})

//提交修改
const handleUpdate = async () => {
    if (!title.value || !content.value) {
        alert('标题和内容不能为空')
        return
    }

    try {
        //发送PUT请求，修改文章
        await request.put(`/v1/articles/${articleId}`, {
            title: title.value,
            content: content.value
        })
        alert('修改成功！')
        router.push('/articles')
    } catch (err) {
        alert('修改失败：' + (err.response?.data?.error || '网络错误'))
    }
}
</script>

<template>
    <div class="edit-box">
        <h2>编辑文章</h2>
        <div v-if="isLoading">正在加载原文章数据...</div>
        <div v-else class="form-group">
            <input v-model="title" type="text" placeholder="请输入文章标题" />
            <textarea v-model="content" rows="8" placeholder="请输入文章内容"></textarea>

            <div class="btn-group">
                <button @click="handleUpdate" class="save-btn">保存修改</button>
                <button @click="$router.push('/articles')" class="cancel-btn">取消</button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.edit-box {
    width: 600px;
    margin: 50px auto;
    padding: 20px;
    border: 1px solid #ccc;
    border-radius: 8px;
}

.form-group {
    display: flex;
    flex-direction: column;
    gap: 15px;
}

input,
textarea {
    padding: 10px;
    font-size: 16px;
    border: 1px solid #ddd;
    border-radius: 4px;
}

.btn-group {
    display: flex;
    gap: 10px;
}

button {
    padding: 10px 20px;
    font-size: 16px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.save-btn {
    background-color: #42b883;
    color: white;
}

.cancel-btn {
    background-color: #ccc;
    color: #333;
}
</style>