const { defineComponent } = Vue;

window.BlogList = defineComponent({
  name: "BlogList",
  props: {
    blogs: {
      type: Array,
      required: true,
    },
    drafts: {
      type: Object,
      required: true,
    },
  },
  emits: ["submit-comment", "update-draft"],
  methods: {
    updateDraft(blogId, event) {
      this.$emit("update-draft", {
        blogId,
        value: event.target.value,
      });
    },
    submit(blogId) {
      this.$emit("submit-comment", blogId);
    },
  },
  template: `
    <section>
      <div class="card blog" v-for="blog in blogs" :key="blog.id">
        <h3>{{ blog.title }}</h3>
        <div class="blog-meta">{{ blog.author }} · {{ blog.createdAt }}</div>
        <p>{{ blog.content }}</p>

        <div>
          <strong>评论</strong>
          <div v-if="blog.comments.length === 0" class="comment">暂无评论，快来抢沙发！</div>
          <div v-for="comment in blog.comments" :key="comment.id" class="comment">
            <div class="comment-meta">{{ comment.author }} · {{ comment.createdAt }}</div>
            <div>{{ comment.content }}</div>
          </div>
        </div>

        <div class="comment-form">
          <input
            :value="drafts[blog.id]"
            placeholder="写下你的评论"
            @input="updateDraft(blog.id, $event)"
          />
          <button @click="submit(blog.id)">发表评论</button>
        </div>
      </div>
    </section>
  `,
});
