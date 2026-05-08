<script setup lang="ts">
import { ref } from "vue";

const query = ref("");
</script>

<template>
  <form class="search-form" action="/search" method="get">
    <input v-model="query" name="q" />
    <button class="primary" type="submit">搜索</button>
  </form>
</template>
