import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
    const token = ref('')
    const username = ref('')
    function setAuth(newToken, newUsername) {
        console.log("正在存入的用户名:", newUsername)
        token.value = newToken
        username.value = newUsername
    }

    function logout() {
        token.value = ''
        username.value = ''
    }

    return  {token, username, setAuth, logout }
}, {
    // 启用持久化插件
    persist: true
})