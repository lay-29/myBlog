const { defineComponent } = Vue;

window.ProfileCard = defineComponent({
  name: "ProfileCard",
  props: {
    profile: {
      type: Object,
      required: true,
    },
    auth: {
      type: Object,
      required: true,
    },
    authMessage: {
      type: String,
      default: "",
    },
  },
  emits: ["register", "login", "update-auth"],
  methods: {
    updateField(field, event) {
      this.$emit("update-auth", { ...this.auth, [field]: event.target.value });
    },
  },
  template: `
    <section class="card profile">
      <img
        src="https://images.unsplash.com/photo-1544723795-3fb6469f5b39?auto=format&fit=crop&w=200&q=80"
        alt="avatar"
      />
      <h2>{{ profile.name }}</h2>
      <p>{{ profile.intro }}</p>
      <div>
        <span class="tag" v-for="tag in profile.tags" :key="tag">{{ tag }}</span>
      </div>

      <div class="auth-section" style="margin-top: 24px;">
        <h3>用户注册 / 登录</h3>
        <input
          :value="auth.username"
          placeholder="用户名"
          @input="updateField('username', $event)"
        />
        <input
          :value="auth.password"
          type="password"
          placeholder="密码"
          @input="updateField('password', $event)"
        />
        <button @click="$emit('register')">注册</button>
        <button @click="$emit('login')">登录</button>
        <div class="auth-message" v-if="authMessage">{{ authMessage }}</div>
      </div>
    </section>
  `,
});
