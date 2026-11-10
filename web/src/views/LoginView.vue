<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import request from '@/utils/request'
const router = useRouter()
const userStore = useUserStore()
const username = ref('')
const password = ref('')

// async function handleLogin() {
//     try {
//         const res = await request.post('/v1/login',{
//             username: username.value,
//             password: password.value
//         })

//         userStore.setAuth(res.token, res.user.username)
//         router.push('/articles')
//     } catch (err) {
//         alert('登录失败: ' + (err.response?.data?.error || '未知错误'))
//     }
// }

const handleLogin = async () => {
    if (!username.value || !password.value) {
        alert('账号密码不能为空')
        return
    }

    try {
        const res = await request.post('/v1/login',{
            username:username.value,
            password:password.value
        })

        console.log('后端返回的数据：', res)

        userStore.setAuth(res.token, res.user.username)
        alert('登录成功')
        router.push('/articles')
    } catch (err) {
        console.error('登录报错:', err)
        alert('登录失败：' + (err.response?.data?.error || '网络错误'))
    }
}
</script>

<template>
    <div class = "login-box">
        <h2>登录 MiniBlog</h2>
            <input v-model="username" type = "text" placeholder = "请输入用户名" />
            <input v-model="password" type = "password" placeholder = "请输入密码"/>

            <button @click = "handleLogin">登录</button>
            <p class="toggle-text" @click="router.push('/register')">
            没有账号？去注册
            </p>
    </div>
</template>

<style scoped>
.login-box {
  width: 300px;
  margin: 100px auto; /* 居中显示 */
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 15px; /* 元素之间的间距 */
}
.login-box h2 {
  text-align: center;
  margin: 0;
}
.login-box input, .login-box button {
  padding: 10px;
  font-size: 16px;
}
.login-box button {
  background-color: #42b883; /* Vue 的经典绿色 */
  color: white;
  border: none;
  cursor: pointer;
  border-radius: 4px;
}
.login-box button:hover {
  background-color: #33a06f;
}
.toggle-text {
  text-align: center;
  color: #666;
  cursor: pointer;
  font-size: 14px;
  margin: 0;
}
.toggle-text:hover {
  color: #42b883;
}
</style>