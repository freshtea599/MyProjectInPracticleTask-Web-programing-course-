import { createRouter, createWebHistory } from 'vue-router'
import CoffeView from '../views/Coffe.vue'
import TodoView from '../views/Todo.vue'
import AboutView from '../views/About.vue'
import RateView from '../views/Rate.vue'
import PersonView from '../views/Person.vue'
import LoginView from '../views/Login.vue'
import RegisterView from '../views/Register.vue'
import ProfileView from '../views/Profile.vue'
import ProductsView from '../views/Products.vue'
import CartView from '../views/Cart.vue'
import AdminProducts from '../views/AdminProducts.vue'
import AdminReviews from '../views/AdminReviews.vue'

const routes = [
  { path: '/', name: 'coffe', component: CoffeView },
  { path: '/about', name: 'about', component: AboutView },
  { path: '/person/:personId', name: 'person', component: PersonView },
  { path: '/rate', name: 'rate', component: RateView },
  { path: '/todo', name: 'todo', component: TodoView, meta: { requiresAuth: true } },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/register', name: 'register', component: RegisterView },
  { path: '/profile', name: 'profile', component: ProfileView, meta: { requiresAuth: true } },
  { path: '/products', name: 'products', component: ProductsView },
  { path: '/cart', component: CartView, meta: { requiresAuth: true } },
  { path: '/admin/products', component: AdminProducts, meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/admin/reviews', component: AdminReviews, meta: { requiresAuth: true, requiresAdmin: true } },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const role = localStorage.getItem('role')

  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.meta.requiresAdmin && role !== 'admin') {
    next('/')
  } else {
    next()
  }
})

export default router
