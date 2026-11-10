<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { useUserStore } from '@/stores/user' 

const router = useRouter()
const userStore = useUserStore() 
const articles = ref([])
const isLoading = ref(false)     

async function loadArticles() {
    isLoading.value = true
    try {
        const res = await request.get('/v1/articles', {
            params: { page: 1, page_size: 10 }
        })
        console.log("获取到的数据:", res) // 加上这行，方便你看数据结构
        articles.value = res.data 
    } catch (err) {
        alert('获取列表失败: ' + (err.response?.data?.error || '网络错误'))
    } finally {
        isLoading.value = false
    }
}

onMounted(loadArticles)

const deleteArticle = async (id) => {
    if (!confirm('确定要删除这篇文章吗？删除后不可恢复！')) return;
    try {
        // 👇 注意这里使用的是反引号 ` 
        await request.delete(`/v1/articles/${id}`);
        alert('删除成功');
        loadArticles();
    } catch (err) {
        alert('删除失败:' + (err.response?.data?.error || '网络错误'));
    }
}
</script>

<template>
  <div style="padding: 20px; max-width: 800px; margin: 0 auto;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
        <h2>文章列表</h2>
        <button @click="$router.push('/articles/create')" style="padding: 8px 16px; cursor: pointer; background: #42b883; color: white; border: none; border-radius: 4px;">
            + 发布新文章
        </button>
    </div>

    <div v-if="isLoading" style="text-align: center; color: #666; margin-top: 50px;">
      
        正在加载文章...
    </div>

    <div v-else-if="articles.length === 0" style="text-align: center; color: #666; margin-top: 50px;">
        暂无文章，快去发布一篇吧！
    </div>

    <div v-else>
        <div v-for="article in articles" :key="article.ID" style="border: 1px solid #eee; border-radius: 8px; padding: 15px; margin-bottom: 15px; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">
    
            <h3 style="margin-top: 0; color: #333;">{{ article.Title || article.title }}</h3>
            <p style="color: #666; font-size: 14px;">
                作者：{{ article.User?.Username || article.user?.username || '未知' }} | 
                发布时间：{{ new Date(article.CreatedAt).toLocaleString() }}
            </p>
            <p style="line-height: 1.6;">{{ article.Content || article.content }}</p>
            
            <!-- 只有作者本人才能看到操作按钮 -->
<div style="text-align: right; margin-top: 10px;">
    <button 
        v-if="article.user?.username === userStore.username"
        @click="$router.push(`/articles/${article.ID}/edit`)" 
        style="color: #1890ff; background: none; border: 1px solid #1890ff; padding: 4px 10px; border-radius: 4px; cursor: pointer; margin-right: 10px;"
    >
        编辑
    </button>

    <button 
        v-if="article.user?.username === userStore.username"
        @click="deleteArticle(article.ID)"
        style="color: #ff4d4f; background: none; border: 1px solid #ff4d4f; padding: 4px 10px; border-radius: 4px; cursor: pointer;"
    >
        删除
    </button>
</div>
        </div>
    </div>
    
  </div>
  
  
</template>