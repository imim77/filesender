<script lang="ts">
  import { onMount } from 'svelte';
  import { PeerManager } from '../services/peerManager';
  import {
    SignalingConnection,
    type ClientInfo,
    type WsServerMessage,
  } from '../services/signaling';
  import type { Peer } from '../services/webrtc';

  let signaling: SignalingConnection | null = null;
  let peerManager: PeerManager | null = null;

  let me: ClientInfo | null = $state(null);
  let peers: ClientInfo[] = $state([]);
  let sessions: Peer[] = $state([]);
  let lastError = $state('');

  function shouldInitiate(peerId: string): boolean {
    if (!me) return false;
    return me.id.localeCompare(peerId) < 0;
  }

  function refreshSessions() {
    sessions = peerManager ? Array.from(peerManager.peersBySessionId.values()) : [];
  }

  function upsertPeer(peer: ClientInfo) {
    if (me && peer.id === me.id) return;

    const index = peers.findIndex((entry) => entry.id === peer.id);
    if (index < 0) {
      peers = [...peers, peer];
      return;
    }

    const next = peers.slice();
    next[index] = peer;
    peers = next;
  }

  function removePeer(peerId: string) {
    peers = peers.filter((peer) => peer.id !== peerId);
  }

  async function handleServerMessage(msg: WsServerMessage) {
    console.log('[WS] incoming:', msg.type, msg);
    await peerManager?.handleMessage(msg);
    refreshSessions();

    switch (msg.type) {
      case 'HELLO':
        me = msg.client;
        peers = msg.peers.filter((peer) => peer.id !== msg.client.id);
        for (const peer of peers) {
          connectToPeer(peer.id, true);
        }
        break;
      case 'JOIN':
        upsertPeer(msg.peer);
        connectToPeer(msg.peer.id, true);
        break;
      case 'UPDATE':
        upsertPeer(msg.peer);
        break;
      case 'LEFT':
        removePeer(msg.peerId);
        break;
      default:
        break;
    }
  }

  function connectToPeer(peerId: string, isAutomatic = false) {
    if (!peerManager) return;

    if (isAutomatic && !shouldInitiate(peerId)) {
      console.log('[AUTO CONNECT] skipping (wait for remote offer):', peerId);
      return;
    }

    const alreadyConnected = sessions.some((session) => session.peerId === peerId);
    if (alreadyConnected) {
      console.log('[CONNECT] session already exists:', peerId);
      return;
    }

    console.log(isAutomatic ? '[AUTO CONNECT] starting session to:' : '[CONNECT] starting session to:', peerId);
    peerManager.startSession(peerId);
    refreshSessions();
  }

  function sendPing(session: Peer) {
    if (session.dc?.readyState !== 'open') return;
    session.dc.send(`ping:${Date.now()}`);
  }

  onMount(() => {
    signaling = new SignalingConnection({
      info: { alias: `Browser-${Math.floor(Math.random() * 1000)}`, deviceType: 'Browser' },
      onOpen: () => {
        console.log('[WS] connected to signaling server');
      },
      onMessage: async (msg) => {
        try {
          await handleServerMessage(msg);
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
      onPeerCreated: refreshSessions,
      onPeerRemoved: refreshSessions,
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
  <p>Me: {me ? `${me.alias || 'Anonymous'} (${me.id})` : 'Connecting...'}</p>
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
          <button onclick={() => connectToPeer(peer.id)}>Connect</button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<section>
  <h2>Sessions</h2>
  {#if sessions.length === 0}
    <p>No active sessions.</p>
  {:else}
    <ul>
      {#each sessions as session}
        <li>
          <span>{session.peerId}</span>
          <span>state: {session.pc?.connectionState ?? 'new'}</span>
          <span>dc: {session.dc?.readyState ?? 'none'}</span>
          <button onclick={() => sendPing(session)} disabled={session.dc?.readyState !== 'open'}>
            Send Ping
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

{#if lastError}
  <p>Last error: {lastError}</p>
{/if}
