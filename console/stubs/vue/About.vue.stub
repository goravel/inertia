<template>
  <Head :title="title" />
  <div class="about-page">
    <header class="head">
      <h1 class="page-title">{{ title }}</h1>
      <p class="lead">{{ content }} Frontend: Vue 3.</p>
    </header>

    <section class="card">
      <h2 class="sub">Stack</h2>
      <ul class="stack">
        <li v-for="dep in deps" :key="dep.href">
          <a class="dep" :href="dep.href" target="_blank" rel="noopener noreferrer">
            <span class="tag" :class="dep.tag">{{ dep.name }}</span>
            <span class="dep-text">{{ dep.text }}</span>
            <span class="ext" aria-hidden="true">↗</span>
          </a>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Head } from '@inertiajs/vue3'

defineProps<{
  title: string
  content: string
}>()

const deps = [
  { name: 'Goravel', tag: 'go', text: 'The Go web framework', href: 'https://github.com/goravel/framework' },
  { name: 'goravel-inertia', tag: 'adapter', text: 'This Inertia adapter', href: 'https://github.com/goravel/inertia' },
  { name: 'Inertia.js', tag: 'inertia', text: 'v3 protocol + Vue 3 adapter', href: 'https://github.com/inertiajs/inertia' },
  { name: 'Vue', tag: 'vue3', text: 'The UI framework', href: 'https://github.com/vuejs/core' },
  { name: 'petaki/inertia-go', tag: 'petaki', text: 'Server engine', href: 'https://github.com/petaki/inertia-go' },
  { name: 'Vite', tag: 'vite', text: 'Dev server + production build', href: 'https://github.com/vitejs/vite' },
]
</script>

<style scoped>
.head {
  margin-bottom: 1.5rem;
}
.sub {
  font-size: 1.1rem;
  font-weight: 700;
  margin-bottom: 1rem;
}
.stack {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.dep {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 10px;
  text-decoration: none;
  color: var(--muted);
  font-size: 0.95rem;
  transition: border-color 0.15s, background 0.15s;
}
.dep:hover {
  border-color: var(--go);
  background: #f7fbfd;
}
.dep-text {
  flex: 1;
}
.ext {
  color: #b3b9c4;
  font-weight: 700;
}
.dep:hover .ext {
  color: var(--go);
}
.tag {
  font-size: 0.78rem;
  font-weight: 700;
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  color: #fff;
}
.tag.go {
  background: #00add8;
}
.tag.adapter {
  background: linear-gradient(135deg, #00add8, #8b5cf6);
}
.tag.inertia {
  background: #8b5cf6;
}
.tag.vue3 {
  background: #42b883;
}
.tag.petaki {
  background: #1f2937;
}
.tag.vite {
  background: #ffae00;
  color: #1f2937;
}
</style>
