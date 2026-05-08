<script setup lang="ts">
import { computed, ref } from "vue";

const mode = ref<"markdown" | "rich">("markdown");
const markdown = ref("");
const richText = ref("");

const content = computed(() => (mode.value === "markdown" ? markdown.value : richText.value));
</script>

<template>
  <section class="stack-form">
    <div class="segmented">
      <button type="button" :class="{ active: mode === 'markdown' }" @click="mode = 'markdown'">Markdown</button>
      <button type="button" :class="{ active: mode === 'rich' }" @click="mode = 'rich'">富文本</button>
    </div>
    <textarea v-if="mode === 'markdown'" v-model="markdown" rows="14" />
    <div v-else class="rich-editor" contenteditable="true" @input="richText = ($event.target as HTMLElement).innerText" />
    <input type="hidden" name="content_markdown" :value="content" />
  </section>
</template>
