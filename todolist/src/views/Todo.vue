<template>
  <div class="row py-3">
    <!-- Панель поиска и сортировки -->
    <div class="col-12 mb-3 d-flex gap-2">
      <input
        v-model="searchQuery"
        type="text"
        class="form-control"
        placeholder="Поиск по группам и задачам..."
      />
      <select v-model="sortOption" class="form-select" style="max-width: 200px;">
        <option value="date">По дате (новые сверху)</option>
        <option value="alpha">По алфавиту</option>
        <option value="unfinished">Невыполненные сверху</option>
      </select>
    </div>

    <!-- Левая колонка: группы -->
    <div class="col-4 border-end">
      <h5 class="d-flex justify-content-between">
        Группы
        <button class="btn btn-sm btn-primary" @click="addGroupVisible = !addGroupVisible">
          <i class="bi bi-plus-lg"></i>
        </button>
      </h5>

      <div v-if="addGroupVisible" class="input-group mb-2">
        <input v-model="newGroupTitle" type="text" class="form-control" placeholder="Название группы" @keyup.enter="addGroup" />
        <button class="btn btn-success" @click="addGroup"><i class="bi bi-check-lg"></i></button>
      </div>

      <ul class="list-group">
        <li
          v-for="group in filteredAndSortedGroups"
          :key="group.id"
          @click="selectGroup(group.id)"
          class="list-group-item d-flex justify-content-between align-items-center"
          :class="{ active: group.id === selectedGroupId }"
          style="cursor: pointer;"
        >
          <span>{{ group.title }}</span>
          <div>
            <i class="bi bi-pencil me-2 text-primary" @click.stop="editGroup(group)"></i>
            <i class="bi bi-trash text-danger" @click.stop="deleteGroup(group.id)"></i>
          </div>
        </li>
      </ul>
    </div>

    <!-- Правая колонка: задачи -->
    <div class="col-8">
      <template v-if="selectedGroup">
        <h5 class="d-flex justify-content-between">
          {{ selectedGroup.title }}
          <button class="btn btn-sm btn-primary" @click="addTaskVisible = !addTaskVisible">
            <i class="bi bi-plus-lg"></i>
          </button>
        </h5>

        <div v-if="addTaskVisible" class="input-group mb-2">
          <input v-model="newTaskTitle" type="text" class="form-control" placeholder="Название задачи" @keyup.enter="addTask" />
          <button class="btn btn-success" @click="addTask"><i class="bi bi-check-lg"></i></button>
        </div>

        <ul class="list-group">
          <li
            v-for="task in filteredAndSortedTasks"
            :key="task.id"
            class="list-group-item d-flex justify-content-between align-items-center"
          >
            <div @click="toggleTask(task.id)" style="cursor: pointer;">
              <i :class="task.done ? 'bi bi-check-circle text-success' : 'bi bi-circle text-secondary'"></i>
              <span :class="{ 'text-decoration-line-through text-muted': task.done }" class="ms-2">{{ task.title }}</span>
            </div>
            <div>
              <i class="bi bi-pencil me-2 text-primary" @click.stop="editTask(task)"></i>
              <i class="bi bi-trash text-danger" @click.stop="deleteTask(task.id)"></i>
            </div>
          </li>
        </ul>
      </template>


      <template v-else>
      </template>
    </div>
  </div>
</template>
<script>

export default {
  data() {
    return {
      groups: JSON.parse(localStorage.getItem("todo-groups") || "[]"),
      selectedGroupId: null,
      searchQuery: "",
      sortOption: "date",

      addGroupVisible: false,
      addTaskVisible: false,

      newGroupTitle: "",
      newTaskTitle: ""
      
    };
  },

  computed: {
    selectedGroup() {
      return this.groups.find(g => g.id === this.selectedGroupId) || null;
    },

    filteredAndSortedGroups() {
      let output = [...this.groups];

      if (this.searchQuery.trim()) {
        const q = this.searchQuery.toLowerCase();
        output = output.filter(g => g.title.toLowerCase().includes(q));
      }

      output = this.applySort(output, "group");
      return output;
    },

    filteredAndSortedTasks() {
      if (!this.selectedGroup) return [];
      let output = [...this.selectedGroup.tasks];

      if (this.searchQuery.trim()) {
        const q = this.searchQuery.toLowerCase();
        output = output.filter(t => t.title.toLowerCase().includes(q));
      }

      output = this.applySort(output, "task");
      return output;
    }
  },

  methods: {
    saveToStorage() {
      localStorage.setItem("todo-groups", JSON.stringify(this.groups));
    },

    selectGroup(id) {
      this.selectedGroupId = id;
      this.addTaskVisible = false;
    },

    addGroup() {
      if (!this.newGroupTitle.trim()) return;
      this.groups.unshift({
        id: Date.now(),
        title: this.newGroupTitle.trim(),
        tasks: []
      });
      this.newGroupTitle = "";
      this.addGroupVisible = false;
      this.saveToStorage();
    },

    editGroup(group) {
      const newTitle = prompt("Введите новое название группы", group.title);
      if (newTitle && newTitle.trim()) {
        group.title = newTitle.trim();
        this.saveToStorage();
      }
    },

    deleteGroup(id) {
      if (confirm("Удалить группу?")) {
        this.groups = this.groups.filter(g => g.id !== id);
        if (this.selectedGroupId === id) this.selectedGroupId = null;
        this.saveToStorage();
      }
    },

    addTask() {
      if (!this.selectedGroup || !this.newTaskTitle.trim()) return;
      this.selectedGroup.tasks.unshift({
        id: Date.now(),
        title: this.newTaskTitle.trim(),
        done: false
      });
      this.newTaskTitle = "";
      this.addTaskVisible = false;
      this.saveToStorage();
    },

    editTask(task) {
      const newTitle = prompt("Введите новое название задачи", task.title);
      if (newTitle && newTitle.trim()) {
        task.title = newTitle.trim();
        this.saveToStorage();
      }
    },

    toggleTask(id) {
      const task = this.selectedGroup.tasks.find(t => t.id === id);
      if (task) {
        task.done = !task.done;
        this.saveToStorage();
      }
    },

    deleteTask(id) {
      if (confirm("Удалить задачу?")) {
        this.selectedGroup.tasks = this.selectedGroup.tasks.filter(t => t.id !== id);
        this.saveToStorage();
      }
    },

    applySort(list, type) {
      if (this.sortOption === "alpha") {
        return list.sort((a, b) => a.title.localeCompare(b.title));
      }
      if (this.sortOption === "unfinished" && type === "task") {
        return list.sort((a, b) => (a.done === b.done ? 0 : a.done ? 1 : -1));
      }
      return list.sort((a, b) => b.id - a.id); 
    }
  }
};
</script>


<style scoped>
.list-group-item.active {
  background-color: #0d6dfd2a;
  border-color: #0d6dfd38;
  color: #fff;
}
</style>
