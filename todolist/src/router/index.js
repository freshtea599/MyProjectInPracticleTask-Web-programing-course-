import { createRouter, createWebHistory } from 'vue-router'
import CoffeView from '../views/Coffe.vue'
import TodoView from '../views/Todo.vue'
import AboutView from '../views/About.vue'
import RateView from '../views/Rate.vue'
import PersonView from '../views/Person.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'coffe', component: CoffeView },
    { path: '/todo', name: 'todo', component: TodoView },
    { path: '/about', name: 'about', component: AboutView },
    { path: '/person', name: 'person', component: PersonView },
    { path: '/rate', name: 'rate', component: RateView },
  ],
})

export default router
