<script lang="ts">
    import { onMount } from "svelte";
    import { SignalingConnection } from "../services/signaling";

  let count: number = $state(0)
  const increment = () => {
    count += 1
  }
  let signaling: SignalingConnection
  onMount(()=>{
    signaling = new SignalingConnection({
      info: { alias: 'TestUser', deviceType: 'Browser' },
      onOpen: () => {
        console.log('Connected to signaling server');
      },
      onError: (e) => {
        console.error('Signaling error:', e);
      },
    })

    return () => {
      signaling?.destroy();
    };
  })
</script>

<button onclick={increment}>
  count is {count}
</button>

<button onclick={() => signaling?.send({ type: 'UPDATE', info: { alias: 'ClickedUser' } })}>
  Send Test UPDATE
</button>
