import {createRouter, createWebHashHistory} from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ChatView from '../views/ChatView.vue'
import authService from '../services/authService.js'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			redirect: '/chat'
		},
		{
			path: '/login',
			name: 'Login',
			component: LoginView,
			meta: { requiresGuest: true }
		},
		{
			path: '/chat',
			name: 'Chat',
			component: ChatView,
			meta: { requiresAuth: true }
		},
		{
			path: '/chat/:id',
			name: 'ChatWithConversation',
			component: ChatView,
			meta: { requiresAuth: true },
			props: true
		},
		// Legacy redirects for backwards compatibility
		{
			path: '/dashboard',
			redirect: '/chat'
		},
		{
			path: '/conversation/:id',
			redirect: to => `/chat/${to.params.id}`
		}
	]
})

// Navigation guards
router.beforeEach((to, _from, next) => {
	const isAuthenticated = authService.isAuthenticated()
	
	if (to.meta.requiresAuth && !isAuthenticated) {
		next('/login')
	} else if (to.meta.requiresGuest && isAuthenticated) {
		next('/chat')
	} else {
		next()
	}
})

export default router
