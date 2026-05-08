<script setup lang="ts">
import { ref } from "vue";

const username = ref("");
const role = ref("member");
</script>

<template>
  <form class="stack-form">
    <label>用户名 <input v-model="username" name="username" /></label>
    <label>角色
      <select v-model="role" name="role">
        <option value="member">成员</option>
        <option value="admin">管理员</option>
        <option value="owner">所有者</option>
      </select>
    </label>
  </form>
</template>
