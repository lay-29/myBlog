const { defineComponent } = Vue;

window.SiteInfo = defineComponent({
  name: "SiteInfo",
  props: {
    blogsLength: {
      type: Number,
      required: true,
    },
    totalComments: {
      type: Number,
      required: true,
    },
  },
  template: `
    <aside class="card site-info">
      <h3>站点信息</h3>
      <ul>
        <li>文章数：{{ blogsLength }}</li>
        <li>评论数：{{ totalComments }}</li>
        <li>上线时间：2024-07</li>
        <li>技术栈：Vue 3 + Go + Gin</li>
      </ul>
      <div style="margin-top: 16px;">
        <h4>快速入口</h4>
        <ul>
          <li>项目介绍</li>
          <li>留言板</li>
          <li>友情链接</li>
        </ul>
      </div>
    </aside>
  `,
});
