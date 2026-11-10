import axios from 'axios'
import {useUserStore } from '@/stores/user'

const request = axios.create({
    baseURL: '/api',
    timeout: 10000
})
request.interceptors.request.use(config => {
    const userStore = useUserStore()
    if (userStore.token) {
        config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
})
// 响应拦截器：直接返回 data，处理 401
request.interceptors.response.use(
    response => response.data,
    error => {
        if (error.Response?.status === 401) {
            const userStore = useUserStore()
            userStore.logout()
            alert('登录已过期，请重新登录')
            window.location.href = '/login'//强制跳转
        }
        return Promise.reject(error)
    }
)

export default request
