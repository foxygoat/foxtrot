<script>
  import { writable } from "svelte/store"

  import Home from "./Home.svelte"
  import Login from "./Login.svelte"
  import Room from "./Room.svelte"

  const page = writable({})
  $: stale = $page.stale

  // hash =  "#room/$Kitchen"
  async function navigate(hash) {
    const dataURL = getDataURL(hash)
    console.log("navigate", hash, "dataURL", dataURL)
    const props = {}
    if (dataURL) {
      $page = { ...$page, stale: true }
      props.data = await fetchData(dataURL)
    }
    const component = getComponent(hash)
    $page = { stale: false, hash, props, component }
    const url = new URL(window.location)
    url.hash = hash
    window.history.pushState({}, "", url)
  }

  async function fetchData(url) {
    const res = await fetch(url)
    if (!res.ok) throw new Error(res) // TODO: error handling.
    return await res.json()
  }

  // hash example: #room/$Kitchen
  function getDataURL(hash) {
    const API_BASE_URL = "http://localhost:8080/" //TODO
    if (hash.startsWith("#room/")) {
      const room = hash.slice(6)
      return API_BASE_URL + `api/history?room=${room}`
    }
  }
  function getComponent(hash) {
    if (hash.startsWith("#room/")) return Room
    if (hash === "#login") return Login
    return Home
  }

  window.onhashchange = () => navigate(window.location.hash)
  navigate(window.location.hash)
</script>

<div class="layout">
  <header>
    <h1 on:click={() => navigate('')}>🦊💃 <span>Foxtrot</span></h1>
    <button on:click={() => navigate('#login')}>Login</button>
    <button on:click={() => navigate('#room/$Kitchen')}>Kitchen</button>
    <button on:click={() => navigate('#room/$Shed')}>Shed</button>
  </header>
  <main>
    <div class="overlay" class:stale />
    <svelte:component this={$page.component} {...$page.props} />
  </main>
</div>

<style>
  .layout {
    @apply flex flex-col max-w-screen-md mx-auto h-screen border border-green-600;
  }
  span {
    @apply hidden sm:inline;
  }
  header {
    @apply flex-none flex items-center justify-between h-12 md:h-16 w-full px-3 sm:px-12 border-b;
  }
  main {
    @apply flex-grow relative min-w-full p-4 sm:p-12 border border-red-600;
  }
  .overlay {
    @apply absolute inset-0 bg-gray-300 bg-opacity-80;
  }
  .stale {
    transition: background 350ms ease-out;
    @apply bg-opacity-30;
  }
</style>
