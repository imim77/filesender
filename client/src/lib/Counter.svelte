<script lang="ts">
  import { onMount } from 'svelte';
  import { PeerManager } from '../services/peerManager';
  import {
    SignalingConnection,
    type ClientInfo,
  } from '../services/signaling';
  import { generateName, getAgentInfo } from '../utilis/uaNames';

  let signaling: SignalingConnection | null = null;
  let peerManager: PeerManager | null = $state(null);

  let me: ClientInfo | null = $state(null);
  let peers: ClientInfo[] = $state([]);
  let lastError = $state('');
  let selectedFiles: FileList | null = null;
  const localAlias = generateName();

  function onFilesSelected(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    selectedFiles = input.files;
  }

  function sendSelectedFiles() {
    if (!peerManager) {
      console.warn('[UI] peer manager not ready');
      return;
    }

    if (!selectedFiles || selectedFiles.length === 0) {
      console.warn('[UI] no files selected');
      return;
    }

    const result = peerManager.sendFilesToConnectedPeers(selectedFiles);
    if (result.peers === 0) {
      console.warn('[UI] no connected peers to send files to');
      return;
    }

    console.log('[UI] queued files for connected peers', {
      files: result.files,
      peers: result.peers,
    });
  }

  onMount(() => {
    signaling = new SignalingConnection({
      info: { alias: localAlias, deviceType: getAgentInfo(navigator.userAgent) },
      onOpen: () => {
        console.log('[WS] connected to signaling server');
      },
      onMessage: async (msg) => {
        try {
          console.log('[WS] incoming:', msg.type, msg);
          await peerManager?.handleMessage(msg);
          me = peerManager?.getSelf() ?? null;
          peers = peerManager?.getPeers() ?? [];
        } catch (error) {
          lastError = error instanceof Error ? error.message : String(error);
          console.error('Failed to handle WS message:', error);
        }
      },
      onClose: (event) => {
        console.log('[WS] signaling closed:', event.code, event.reason);
      },
      onError: (error) => {
        lastError = error instanceof Error ? error.message : String(error);
        console.error('Signaling error:', error);
      },
    });

    peerManager = new PeerManager({
      signaling,
      onPeerCreated: (peer) => {
        console.log('[PeerManager] peer session created', {
          sessionId: peer.sessionId,
          peerId: peer.peerId,
          isCaller: peer.isCaller,
        });
      },
      onPeerRemoved: (peer) => {
        console.log('[PeerManager] peer session removed', {
          sessionId: peer.sessionId,
          peerId: peer.peerId,
        });
      },
      onError: (error) => {
        lastError = error instanceof Error ? error.message : String(error);
        console.error('Peer manager error:', error);
      },
    });

    return () => {
      peerManager?.destroy();
      signaling?.destroy();
    };
  });
</script>

<section>
  <h2>Signaling</h2>
  <p>Me: {me ? `${me.alias || localAlias} (${me.id})` : `Connecting as ${localAlias}...`}</p>
</section>

<section>
  <h2>Peers</h2>
  {#if peers.length === 0}
    <p>No peers online.</p>
  {:else}
    <ul>
      {#each peers as peer}
        <li>
          <strong>{peer.alias || 'Anonymous'}</strong>
          <code>{peer.id}</code> 
        </li>
      {/each}
    </ul>
  {/if}
</section>

<section>
  <h2>File Transfer</h2>
  <p>Connected peers: {peerManager?.getConnectedPeerCount() ?? 0}</p>
  <input type="file" multiple onchange={onFilesSelected} />
  <button onclick={sendSelectedFiles}>Send Selected Files</button>
</section>

{#if lastError}
  <p>Last error: {lastError}</p>
{/if}
