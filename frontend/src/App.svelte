<script>
  export let name
</script>

<main>
  <h1 class="text-red-400 text-2xl py-8">Hello {name}!</h1>
  <p>
    Visit the
    <a href="https://svelte.dev/tutorial">Svelte tutorial</a>
    to learn how to build Svelte apps.
  </p>
</main>

<style global>
  /* purgecss start ignore */
  @tailwind base;
  @tailwind components;
  /* purgecss end ignore */
  @tailwind utilities;

  main {
    @apply max-w-sm mx-auto;
  }
</style>
