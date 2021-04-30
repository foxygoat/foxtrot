# A Initial page load with hash

=> Parse "navigation state" out of hash
=> Fetch data
=> set stale indicator

2. New data arrives
   => put it in store  
   => render new Component

- DON'T update URL with history.pushState()
  => remove stale indicator

# B UI, things to click on it.

1. call fetch new data.
   => keep old screen/widget view AND show stale/loading indicator

2. new data arrives:
   => put it in store  
   => render new Component
   => update URL with history.pushState() - doesn't trigger onhashchange event (URL change event)
   => remove stale indicator

# C OnHshChange

1. with "onhashchange" event, ie back/forwards button, manually editing URL
   stale indicator
   load new data

2. new data arrives
   => put it in store  
   => render new Component

- DON'T update URL with history.pushState()
  => remove stale indicator

=================
function navigate(navPath "#room/$Kitchen") {
const dataURL = getDataURL(navPath)
let data
if dataURL {
$page = {stale:true}
data = await fetchData(dataURL)
}
$page = {
stale: false,
navPath: navPath,
data,
component: getComponent(navPath),
}
}

page {
navPath: Path,
stale: true|false,
component: component,
data: data,
}

{ onhashchange((e)=> {navigate(e.hash)}) }

<svelte:component this={component()$page.component} data=$page.data />
<button on:click(navigate("#login"))>Login</button>
<button on:click(navigate("#room/$Kitchen"))>GoKitchen</button>
