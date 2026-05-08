import { createApp } from "vue";
import KnowledgeEditor from "./components/KnowledgeEditor.vue";
import TeamMembers from "./components/TeamMembers.vue";
import SearchPanel from "./components/SearchPanel.vue";

const mounts = [
  ["knowledge-editor", KnowledgeEditor],
  ["team-members", TeamMembers],
  ["search-panel", SearchPanel],
] as const;

for (const [id, component] of mounts) {
  const el = document.getElementById(id);
  if (el) createApp(component).mount(el);
}
