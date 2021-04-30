<script>
  async function getMessages(room) {
    const res = await fetch(API_BASE_URL + `api/history?room=${room}`)
    if (!res.ok) throw new Error(res)
    return await res.json()
  }
  let promise = getMessages()
</script>

{#await promise}
  <p>...loading</p>
{:then messages}
  <div class="messages">
    {#each messages.reverse() as { author, content, createdAt }}
      <div class="message">
        <span class="author">{author}</span>
        <span class="createdAt">{createdAt.substring(11, 16)}</span>
        <div>{content}</div>
      </div>
    {/each}
  </div>
{:catch error}
  <p style="red-400">{error.message}</p>
{/await}

<style>
  .messages {
    @apply w-full md:w-2/3;
  }
  .message {
    @apply py-3;
  }
  .author {
    @apply text-gray-800 font-bold text-sm;
  }
  .createdAt {
    @apply text-gray-400 text-xs;
  }
</style>
