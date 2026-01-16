const { createApp } = Vue;

createApp({
  components: {
    ProfileCard: window.ProfileCard,
    BlogList: window.BlogList,
    SiteInfo: window.SiteInfo,
  },
  data() {
    return {
      profile: {
        name: "林霖",
        intro: "全栈开发者，喜欢记录技术与思考。",
        tags: ["前端", "后端", "Go", "Vue", "生活"],
      },
      auth: {
        username: "",
        password: "",
      },
      authMessage: "",
      blogs: [],
      commentDrafts: {},
    };
  },
  computed: {
    totalComments() {
      return this.blogs.reduce((sum, blog) => sum + blog.comments.length, 0);
    },
  },
  methods: {
    async fetchBlogs() {
      const response = await fetch("/api/blogs");
      this.blogs = await response.json();
    },
    async register() {
      const response = await fetch("/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(this.auth),
      });
      const data = await response.json();
      this.authMessage = data.message || data.error;
    },
    async login() {
      const response = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(this.auth),
      });
      const data = await response.json();
      this.authMessage = data.message || data.error;
    },
    async submitComment(blogID) {
      const content = this.commentDrafts[blogID];
      if (!content) {
        this.authMessage = "评论内容不能为空";
        return;
      }
      const response = await fetch(`/api/blogs/${blogID}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          author: this.auth.username || "匿名",
          content,
        }),
      });
      const data = await response.json();
      if (response.ok) {
        this.blogs = this.blogs.map((blog) =>
          blog.id === blogID ? data : blog
        );
        this.commentDrafts = {
          ...this.commentDrafts,
          [blogID]: "",
        };
        this.authMessage = "评论已发布";
      } else {
        this.authMessage = data.error || "评论失败";
      }
    },
    updateAuth(nextAuth) {
      this.auth = nextAuth;
    },
    updateDraft({ blogId, value }) {
      this.commentDrafts = {
        ...this.commentDrafts,
        [blogId]: value,
      };
    },
  },
  mounted() {
    this.fetchBlogs();
  },
  template: `
    <div>
      <header>
        <h1>我的博客</h1>
        <p>Vue + Go 的轻量博客系统，分享技术与生活。</p>
      </header>

      <main class="layout">
        <ProfileCard
          :profile="profile"
          :auth="auth"
          :auth-message="authMessage"
          @register="register"
          @login="login"
          @update-auth="updateAuth"
        />

        <BlogList
          :blogs="blogs"
          :drafts="commentDrafts"
          @submit-comment="submitComment"
          @update-draft="updateDraft"
        />

        <SiteInfo :blogs-length="blogs.length" :total-comments="totalComments" />
      </main>
    </div>
  `,
}).mount("#app");
