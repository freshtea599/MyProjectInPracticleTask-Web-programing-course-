<script setup>
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { ref, computed } from 'vue'

const router = useRouter()

const loggedIn = computed(() => !!localStorage.getItem('token'))

const isAdminStorage = ref(localStorage.getItem('is_admin')) 


const isAdmin = computed(() => isAdminStorage.value === 'true')

function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('user_id')
  localStorage.removeItem('username')
  localStorage.removeItem('is_admin') 
 
  isAdminStorage.value = null
  
  router.push('/login')
}
</script>

<template>
  <header>
    <nav class="navbar navbar-light bg-light d-flex justify-content-between align-items-center px-3 shadow-sm main-navbar">
      <div class="d-flex gap-3">
        <RouterLink class="navbar-brand" to="/">Айти кофейня</RouterLink>
        <RouterLink class="nav-link" to="/">Главная</RouterLink>
        <RouterLink class="nav-link" to="/rate">Оценить</RouterLink>
        <RouterLink class="nav-link" to="/about">Об авторе</RouterLink>
        <RouterLink v-if="loggedIn" class="nav-link" to="/todo">Заметочки</RouterLink>
        <RouterLink v-if="loggedIn" class="nav-link" to="/products">Магазин</RouterLink>
        <RouterLink v-if="loggedIn" class="nav-link" to="/cart">Корзина</RouterLink>
      </div>
      <div>
        <RouterLink
          v-if="!loggedIn"
          class="btn btn-outline-primary me-2"
          to="/login"
        >
          Войти
        </RouterLink>
        <RouterLink
          v-if="!loggedIn"
          class="btn btn-success me-2"
          to="/register"
        >
          Регистрация
        </RouterLink>
        <div v-if="loggedIn" class="dropdown d-inline">
          <button
            class="btn btn-outline-secondary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
          >
            <i class="bi bi-person-circle"></i>
          </button>
          <ul class="dropdown-menu dropdown-menu-end">
            <li>
              <RouterLink class="dropdown-item" to="/profile">
                <i class="bi bi-person"></i> Профиль
              </RouterLink>
            </li>
            <li>
              <RouterLink class="dropdown-item" to="/todo">
                <i class="bi bi-list-check"></i> Заметки
              </RouterLink>
            </li>
            <li v-if="isAdmin">
              <hr class="dropdown-divider" />
            </li>
            <li v-if="isAdmin">
              <RouterLink class="dropdown-item" to="/admin/products">
                <i class="bi bi-shop"></i> Управление товарами
              </RouterLink>
            </li>
            <li v-if="isAdmin">
              <RouterLink class="dropdown-item" to="/admin/reviews">
                <i class="bi bi-chat-left-text"></i> Модерация отзывов
              </RouterLink>
            </li>
            <li><hr class="dropdown-divider" /></li>
            <li>
              <a class="dropdown-item" @click="logout">
                <i class="bi bi-box-arrow-right"></i> Выйти
              </a>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  </header>

  <main class="container mt-4">
    <RouterView />
  </main>

<footer class="text-center text-muted mt-4 mb-2">
    <small>© 2025 TodoList by Михаил</small>
  </footer>
</template>

<style scoped>
.main-navbar {
  background: linear-gradient(90deg, #f7e2c6, #fbead8);
  border-bottom: 1px solid rgba(0,0,0,0.05);
}

.main-navbar .navbar-brand {
  font-weight: 700;
  letter-spacing: 0.03em;
  color: #5a3c2e;
}

.main-navbar .nav-link {
  color: #6c4a34;
  font-weight: 500;
}

.main-navbar .nav-link.router-link-active {
  color: #3b2418;
  border-bottom: 2px solid #c2874a;
}

.main-navbar .btn-outline-primary,
.main-navbar .btn-success {
  border-radius: 999px;
  padding-inline: 1.4rem;
}

.main-content {
  padding-bottom: 80px;
}
.carousel-item {
  display: none;
  transition: opacity 0.6s ease-in-out;
  opacity: 0;
}

.carousel-item.active {
  display: block;
  opacity: 1;
}

.carousel-indicators .active {
  background-color: #222;
}

.carousel-indicators button {
  background-color: #ddd;
  border: none;
}

.carousel-control-prev-icon,
.carousel-control-next-icon {
  background-color: #000;
  border-radius: 50%;
  padding: 10px;
}

.feedback {
  font-size: 1.8rem;
  margin-top: 2rem;
  margin-bottom: 1.5rem;
}

.reviews-carousel {
  position: relative;
  min-height: 300px;
}

.reviews-carousel .carousel-item {
  padding: 1.5rem;
}

.reviews-carousel .carousel-item p {
  line-height: 1.6;
}


@media (max-width: 575.98px) {
  .feedback {
    font-size: 1.3rem;
    margin-top: 1.5rem;
  }

  .reviews-carousel {
    min-height: 250px;
  }

  .reviews-carousel .carousel-item {
    padding: 1rem;
  }

  .reviews-carousel .carousel-item h5 {
    font-size: 1rem;
  }

  .reviews-carousel .carousel-item p {
    font-size: 0.9rem;
  }

  .reviews-carousel .carousel-item i {
    font-size: 1.5rem;
  }

  .carousel-control-prev,
  .carousel-control-next {
    width: 40px !important;
    height: 40px !important;
  }

  .carousel-control-prev-icon,
  .carousel-control-next-icon {
    padding: 6px;
  }
}


@media (min-width: 576px) and (max-width: 767.98px) {
  .feedback {
    font-size: 1.5rem;
  }

  .reviews-carousel {
    min-height: 280px;
  }

  .reviews-carousel .carousel-item h5 {
    font-size: 1.1rem;
  }
}


@media (min-width: 768px) {
  .feedback {
    font-size: 1.8rem;
  }

  .reviews-carousel {
    min-height: 320px;
  }
}
</style>

