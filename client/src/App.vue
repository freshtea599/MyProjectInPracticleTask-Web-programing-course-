<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <header>
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
      <div class="container-fluid">
        <a class="navbar-brand" href="#">Редактор заметок</a>
        <button class="navbar-toggler" type="button"
        data-bs-toggle="collapse" data-bs-target="#navbarNav"
        aria-controls="navbarNav" aria-expanded="false"
        aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
        </button>
      <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav">
          <li class="nav-item">
            <router-link class="nav-link" to="/">
              <i class="bi-house"></i> Главная</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/rate">
              <i class="bi-star"></i> Оценить</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/about">
              <i class="bi-person-circle"></i> Об авторе</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/todo">
              <i class="bi-house"></i> Заметочки</router-link>
          </li>
        </ul>
      </div>
      </div>
    </nav>
  </header>

  <main class="container mt-4">
    <RouterView />
  </main>

  <footer class="text-center mt-5 mb-3 text-muted">
    <small>© 2025 TodoList by Михаил</small>
  </footer>
</template>
