<script lang="ts">
  import { onDestroy } from 'svelte';
  import { WebRTCController } from './services/webrtccontroller.svelte';
    import { generateName, getAgentInfo } from './utilis/uaNames';

  const controller = new WebRTCController(generateName(), getAgentInfo(navigator.userAgent));

  function connect(peerId: string): void {
    controller.connectToPeer(peerId);
  }

  function sendFiles(peerId: string, event: Event): void {
    const input = event.currentTarget as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;
    controller.sendFiles(peerId, input.files);
    input.value = '';
  }

  function connectionLabel(peerId: string): string {
    return controller.peerConnections.get(peerId)?.isConnected ? 'connected' : 'not connected';
  }

  onDestroy(() => {
    controller.destroy();
  });
</script>

<main>
  <h1>FileSender</h1>
  <p>Status: {controller.connectionStatus}</p>
  <h2>I'am known as {controller.myName}</h2>
  <h2>Peers ({controller.peers.length})</h2>
  {#if controller.peers.length === 0}
    <p>Waiting for peers to join...</p>
  {:else}
    <ul>
      {#each controller.peers as peer}
        <li>
          <div>
            <strong>{peer.alias || 'Unnamed device'}</strong>
            <span>{peer.deviceType || 'Unknown device'}</span>
            <span>{connectionLabel(peer.id)}</span>
          </div>
          <div>
            <button type="button" on:click={() => connect(peer.id)}>Connect</button>
            <input type="file" multiple on:change={(event) => sendFiles(peer.id, event)} />
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: 'Avenir Next', 'Segoe UI', sans-serif;
    background: linear-gradient(160deg, #f2f7f5 0%, #dfeee8 100%);
    color: #1a2a23;
  }

  main {
    max-width: 760px;
    margin: 0 auto;
    padding: 2rem 1rem 3rem;
  }

  h1 {
    margin: 0 0 0.5rem;
  }

  h2 {
    margin-top: 2rem;
  }

  ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    gap: 0.75rem;
  }

  li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem;
    border: 1px solid #b8d1c4;
    border-radius: 10px;
    background: #ffffffcc;
    flex-wrap: wrap;
  }

  li div {
    display: flex;
    gap: 0.6rem;
    align-items: center;
    flex-wrap: wrap;
  }

  button {
    border: 0;
    border-radius: 8px;
    padding: 0.5rem 0.8rem;
    cursor: pointer;
    background: #1d6f50;
    color: white;
  }

  button:hover {
    background: #15553d;
  }

  span {
    font-size: 0.85rem;
    opacity: 0.8;
  }

  @media (max-width: 720px) {
    li {
      align-items: flex-start;
    }
  }
</style>
