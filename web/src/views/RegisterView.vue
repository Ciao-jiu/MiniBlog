<script setup>
import {ref } from 'vue'
import { useRouter } from 'vue-router';
import request from '@/utils/request';
const router = useRouter()
const username = ref('')
const password = ref('')

const handleRegister = async () => {
    if (!username.value || !password.value) {
        alert('账号密码不能为空')
        return 
    }

    try {
        //请求注册接口
        const res = await request.post('/v1/register',{
            username:username.value,
            password:password.value
        })

        console.log('注册成功：',res)
        alert('注册成功！请前往登录')

        router.push('login')
    } catch (err) {
        console.error("注册报错:" , err)
        alert('注册失败:' + (err.response?.data?.error || '网络错误'))
    }
}
</script>

<template>
    <div class="register-box">
        <h2>注册 MiniBlog</h2>
        
        <input v-model="username" type="text" placeholder="设置用户名" />
        <input v-model="password" type="password" placeholder="设置密码（至少6位）" />

        <button @click="handleRegister">注 册</button>
        
        <!-- 提供一个跳回登录页的链接 -->
        <p class="toggle-text" @click="router.push('/login')">
            已有账号？去登录
        </p>
    </div>
</template>

<style scoped>
.register-box {
  width: 300px;
  margin: 100px auto;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 15px;
}
.register-box h2 {
  text-align: center;
  margin: 0;
}
.register-box input, .register-box button {
  padding: 10px;
  font-size: 16px;
}
.register-box button {
  background-color: #42b883;
  color: white;
  border: none;
  cursor: pointer;
  border-radius: 4px;
}
.register-box button:hover {
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