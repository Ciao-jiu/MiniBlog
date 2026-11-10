import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import { useUserStore } from '@/stores/user'
import RegisterView from '@/views/RegisterView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect:'/login' },
    {path:'/login', component: LoginView },
    {path:'/register',component: RegisterView},
    {
      path:'/articles',
      component: () => import('@/views/ArticleListView.vue'),
      meta:{ requiresAuth:true }
    },
    {
      path:'/articles/create',
      component: () => import('@/views/ArticleCreateView.vue'),
      meta: { requiresAuth:true }
    },
    {
      path:'/articles/:id/edit',
      component:() => import('@/views/ArticleEditView.vue'),
      meta: {requiresAuth:true}
    }
  ]
})

router.beforeEach((to, from) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.token){
    return '/login'
  } 
})
export default router
