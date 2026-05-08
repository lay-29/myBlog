(() => {
  const csrf = document.querySelector('meta[name="csrf-token"]')?.content || "";

  document.addEventListener("click", (event) => {
    document.querySelectorAll(".dropdown[open]").forEach((dropdown) => {
      if (!dropdown.contains(event.target)) {
        dropdown.removeAttribute("open");
      }
    });
  });

  document.querySelectorAll(".dropdown-menu a, .dropdown-menu button").forEach((item) => {
    item.addEventListener("click", () => {
      item.closest(".dropdown")?.removeAttribute("open");
    });
  });

  async function postJSON(url, payload) {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrf,
      },
      body: JSON.stringify(payload),
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || "请求失败");
    }
    return data;
  }

  function formJSON(form) {
    const data = new FormData(form);
    return Object.fromEntries(data.entries());
  }

  function bindJSONForm(selector, urlFor, buildPayload) {
    const form = document.querySelector(selector);
    if (!form) return;
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const button = form.querySelector("button[type='submit']");
      const oldText = button?.textContent;
      if (button) {
        button.disabled = true;
        button.textContent = "处理中";
      }
      try {
        const payload = buildPayload ? buildPayload(form) : formJSON(form);
        await postJSON(typeof urlFor === "function" ? urlFor(form) : urlFor, payload);
        window.location.reload();
      } catch (error) {
        alert(error.message);
      } finally {
        if (button) {
          button.disabled = false;
          button.textContent = oldText;
        }
      }
    });
  }

  bindJSONForm("#create-team-form", "/api/v1/teams");
  bindJSONForm("#create-space-form", "/api/v1/spaces", (form) => {
    const payload = formJSON(form);
    payload.team_id = Number(payload.team_id);
    return payload;
  });
  bindJSONForm("#add-member-form", (form) => `/api/v1/teams/${form.dataset.teamId}/members`);

  const markdownEditor = document.querySelector("#markdown-editor");
  const richEditor = document.querySelector("#rich-editor");
  const modeButtons = document.querySelectorAll("[data-editor-mode]");
  function syncRichEditorToMarkdown() {
    if (richEditor && markdownEditor && !richEditor.hidden) {
      markdownEditor.value = richEditor.innerText;
    }
  }

  modeButtons.forEach((button) => {
    button.addEventListener("click", () => {
      modeButtons.forEach((item) => item.classList.remove("active"));
      button.classList.add("active");
      const mode = button.dataset.editorMode;
      if (mode === "rich") {
        richEditor.hidden = false;
        markdownEditor.hidden = true;
        richEditor.innerText = markdownEditor.value;
      } else {
        markdownEditor.hidden = false;
        richEditor.hidden = true;
        markdownEditor.value = richEditor.innerText;
      }
    });
  });

  document.querySelectorAll("[data-editor-form]").forEach((form) => {
    form.addEventListener("submit", syncRichEditorToMarkdown);
  });

  bindJSONForm("#create-article-form", "/api/v1/articles", (form) => {
    syncRichEditorToMarkdown();
    const payload = formJSON(form);
    payload.team_id = Number(payload.team_id);
    payload.space_id = Number(payload.space_id);
    payload.is_blog = payload.is_blog === "true";
    return payload;
  });

  const settingsNavItems = document.querySelectorAll("[data-settings-nav]");
  const settingsPanels = document.querySelectorAll("[data-settings-panel]");
  function activateSettingsPanel(hash) {
    const id = (hash || "#profile").replace("#", "");
    settingsNavItems.forEach((item) => item.classList.toggle("active", item.getAttribute("href") === `#${id}`));
    settingsPanels.forEach((panel) => panel.classList.toggle("active", panel.id === id));
  }
  if (settingsNavItems.length) {
    activateSettingsPanel(window.location.hash);
    settingsNavItems.forEach((item) => {
      item.addEventListener("click", (event) => {
        event.preventDefault();
        const hash = item.getAttribute("href");
        history.replaceState(null, "", hash);
        activateSettingsPanel(hash);
      });
    });
  }
})();
