<template>
<main class="container">
    <h1 class="fw-bold display-5">Айти кофейня - лучшее место для учебы в айти!</h1>
    <p class="lead text-muted">А здесь вы можете узнать информацию о нас</p>
    <div class="row">

      <div class="col-md-4">
        <router-link class="text-decoration-none text-dark" :to="'/person/person-1'">
          <div class="card shadow rounded my-3 employee-card">
            <div class="card-body text-center">
              <div><img src="../assets/image/person-1.jpg" class="avatar"></div>
              <div class="avatar-post"><i class="bi bi-cup-hot"></i> Наш бариста</div>
              <div class="avatar-text">Раньше уничтожал вселенные, а теперь варит кофе</div>
              <p>
                <a href="mailto:barista@itcoffee.ru">barista@itcoffee.ru</a>
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-4">
        <router-link class="text-decoration-none text-dark" :to="'/person/person-2'">
          <div class="card shadow rounded my-3 employee-card">
            <div class="card-body text-center">
              <div><img src="../assets/image/person-2.jpg" class="avatar"></div>
              <div class="avatar-post"><i class="bi bi-lightbulb"></i> Креативный директор</div>
              <div class="avatar-text">Пьет слишком много кофе. Если встретите его — прячьте свой стакан!</div>
              <p>
                <a href="mailto:creative@itcoffee.ru">creative@itcoffee.ru</a>
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-4">
        <router-link class="text-decoration-none text-dark" :to="'/person/person-3'">
          <div class="card shadow rounded my-3 employee-card">
            <div class="card-body text-center">
              <div><img src="../assets/image/person-3.jpg" class="avatar"></div>
              <div class="avatar-post"><i class="bi bi-truck"></i> Логист-бизон</div>
              <div class="avatar-text">Доставит кофе на вертолете. А так у него вообще свой бизнес...</div>
              <p>
                <a href="mailto:logist@itcoffee.ru">logist@itcoffee.ru</a>
              </p>
            </div>
          </div>
        </router-link>
      </div>

    </div>
  </main>
</template>

<script setup>
</script>

<style scoped>
.about-page img {
  width: 150px;
  height: 150px;
  object-fit: cover;
}
</style>
