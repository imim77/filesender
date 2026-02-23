<script lang="ts">
  import { onMount } from 'svelte';
    import { WebRTCController } from '../services/webrtccontroller.svelte';
  let pm = new WebRTCController();


</script>


<main>
  PeerManage:
  {pm}
</main>