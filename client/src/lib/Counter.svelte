<script lang="ts">
    import { onMount } from "svelte";
    import { SignalingConnection } from "../services/signaling";

  let count: number = $state(0)
  const increment = () => {
    count += 1
  }
  let signaling: SignalingConnection
  onMount(()=>{
    signaling = new SignalingConnection({info:{alias:'TestUser', deviceType: 'Browser'}})
    signaling.addEventListener('open', () => {
      console.log('Connected to signaling server');
    });

    signaling.addEventListener('error', (e) => {
      console.error('Signaling error:', e);
    });

    signaling.connect();

    return () => {
      signaling?.socket?.close();
    };
  })
</script>

<button onclick={increment}>
  count is {count}
</button>

<button onclick={() => signaling?.send({ type: 'UPDATE', info: { alias: 'ClickedUser' } })}>
  Send Test UPDATE
</button>
