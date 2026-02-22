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
  const localAlias = generateName();
  let selectedPeerId: string | null = null;
  let fileInput: HTMLInputElement | null = $state(null);

  function openFilePickerForPeer(peerId: string) {
    if (!peerManager) {
      console.warn('[UI] peer manager not ready');
      selectedPeerId = null;
      return;
    }

    if (!peerManager.isPeerConnected(peerId)) {
      console.warn('[UI] selected peer is not connected:', peerId);
      return;
    }

    selectedPeerId = peerId;
    fileInput?.click();
  }

  function onPeerFileSelected(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const files = input.files;

    if (!peerManager) {
      console.warn('[UI] peer manager not ready');
      return;
    }

    if (!selectedPeerId) {
      console.warn('[UI] no target peer selected');
      input.value = '';
      selectedPeerId = null;
      return;
    }

    if (!files || files.length === 0) {
      console.warn('[UI] no files selected');
      input.value = '';
      selectedPeerId = null;
      return;
    }

    const result = peerManager.sendFilesToPeer(selectedPeerId, files);
    if (!result.ok) {
      console.warn('[UI] failed to send files', {
        peerId: selectedPeerId,
        reason: result.reason,
      });
      input.value = '';
      selectedPeerId = null;
      return;
    }

    console.log('[UI] queued files for peer', {
      peerId: selectedPeerId,
      files: result.files,
    });

    input.value = '';
    selectedPeerId = null;
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
    <input bind:this={fileInput} type="file" hidden onchange={onPeerFileSelected} />
    <ul>
      {#each peers as peer}
        <li>
          <strong>{peer.alias || 'Anonymous'}</strong>
          <code>{peer.id}</code>
          <button onclick={() => openFilePickerForPeer(peer.id)} disabled={!peerManager?.isPeerConnected(peer.id)}>
            Send file
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

{#if lastError}
  <p>Last error: {lastError}</p>
{/if}
