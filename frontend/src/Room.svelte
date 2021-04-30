<script>
  import { fade } from "svelte/transition"
  import { beforeUpdate, afterUpdate } from "svelte"

  export let data
  $: messages = data.slice().reverse()
  let div
  let autoscroll

  beforeUpdate(() => {
    autoscroll = div && div.offsetHeight + div.scrollTop > div.scrollHeight - 20
  })

  afterUpdate(() => {
    if (autoscroll) div.scrollTo(0, div.scrollHeight)
  })

  function handleKeydown(event) {
    if (event.key !== "Enter") return
    const text = event.target.value
    if (!text) return

    messages.concat({
      author: "???",
      content: text,
      createdAt: "2020-11-22T11:12:12Z",
    })

    event.target.value = ""
  }
</script>

<div class="room">
  <div class="messages">
    {#each messages as { author, content, createdAt, id } (id)}
      <div class="message" in:fade={{ duration: 250 }}>
        <span class="author">{author}</span>
        <span class="createdAt">{createdAt.substring(11, 16)}</span>
        <div>{content}</div>
      </div>
    {/each}
  </div>
  <input on:keydown={handleKeydown} />
</div>

<style>
  .room {
    @apply flex flex-col w-full md:w-2/3 border border-blue-600 h-full;
  }
  .messages {
    @apply flex-grow w-full md:w-2/3 overflow-y-auto;
  }
  .message {
    @apply py-3; /*TODO: transition here?*/
  }
  .author {
    @apply text-gray-800 font-bold text-sm;
  }
  .createdAt {
    @apply text-gray-400 text-xs;
  }
  input {
    @apply flex-none border;
  }
</style>
