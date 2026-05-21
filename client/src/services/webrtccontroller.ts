import {
  type AnswerMessage,
  type ClientInfo,
  type OfferMessage,
  SignalingConnection,
  type WsServerMessage,
} from './signaling';
import { Peer } from './webrtc';

type PeerStatus = 'connected' | 'not connected';
type ConnectionStatus = 'disconnected' | 'connecting' | 'connected';

export type WebRTCControllerSnapshot = {
  connectionStatus: ConnectionStatus;
  myId: string;
  myName: string;
  peers: ClientInfo[];
  peerConnectionState: Record<string, PeerStatus>;
};

export class WebRTCController {
  private listeners = new Set<() => void>();
  private connectionStatus: ConnectionStatus = 'connecting';
  private myId = '';
  private myName = '';
  private peers: ClientInfo[] = [];
  private peerConnections = new Map<string, Peer>();
  private peerConnectionState: Record<string, PeerStatus> = {};
  private signaling: SignalingConnection;
  private snapshot: WebRTCControllerSnapshot = this.createSnapshot();

  constructor(alias: string, deviceModel: string) {
    this.myName = alias;
    this.signaling = new SignalingConnection({
      info: {
        alias,
        deviceModel,
      },
      onOpen: () => {
        this.connectionStatus = 'connecting';
        this.notify();
      },
      onMessage: (msg) => {
        this.handleSignalingMessage(msg);
      },
      onClose: () => {
        this.connectionStatus = 'disconnected';
        this.cleanupPeerConnections();
        this.peers = [];
        this.peerConnectionState = {};
        this.notify();
      },
      onError: (error) => {
        this.connectionStatus = 'disconnected';
        console.error('[Signaling] connection error', error);
        this.notify();
      },
    });
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  getSnapshot(): WebRTCControllerSnapshot {
    return this.snapshot;
  }

  private createSnapshot(): WebRTCControllerSnapshot {
    return {
      connectionStatus: this.connectionStatus,
      myId: this.myId,
      myName: this.myName,
      peers: [...this.peers],
      peerConnectionState: { ...this.peerConnectionState },
    };
  }

  sendFiles(peerId: string, files: FileList | File[]): void {
    const peer = this.peerConnections.get(peerId);
    if (!peer) return;
    peer.sendFiles(files);
  }

  connectionLabel(peerId: string): PeerStatus {
    return this.peerConnectionState[peerId] ?? 'not connected';
  }

  destroy(): void {
    this.cleanupPeerConnections();
    this.signaling.destroy();
    this.peers = [];
    this.peerConnectionState = {};
    this.connectionStatus = 'disconnected';
    this.notify();
  }

  private notify(): void {
    this.snapshot = this.createSnapshot();
    this.listeners.forEach((listener) => listener());
  }

  private handleSignalingMessage(msg: WsServerMessage): void {
    switch (msg.type) {
      case 'HELLO':
        this.myId = msg.client.id;
        if (msg.client.alias) {
          this.myName = msg.client.alias;
        }
        this.peers = msg.peers.filter((peer) => peer.id !== this.myId);
        this.connectionStatus = 'connected';
        this.peers.forEach((peer) => {
          this.connectPeer(peer.id);
        });
        this.notify();
        return;
      case 'JOIN':
        if (msg.peer.id === this.myId) return;
        if (!this.peers.some((peer) => peer.id === msg.peer.id)) {
          this.peers = [...this.peers, msg.peer];
        }
        this.connectPeer(msg.peer.id);
        this.notify();
        return;
      case 'UPDATE':
        if (msg.peer.id === this.myId) {
          if (msg.peer.alias) {
            this.myName = msg.peer.alias;
          }
          this.notify();
          return;
        }
        this.peers = this.peers.some((peer) => peer.id === msg.peer.id)
          ? this.peers.map((peer) => (peer.id === msg.peer.id ? msg.peer : peer))
          : [...this.peers, msg.peer];
        this.notify();
        return;
      case 'LEFT': {
        const peer = this.peerConnections.get(msg.peerId);
        if (peer) {
          peer.destroy();
          this.peerConnections.delete(msg.peerId);
        }
        this.removePeerConnectionState(msg.peerId);
        this.peers = this.peers.filter((p) => p.id !== msg.peerId);
        this.notify();
        return;
      }
      case 'OFFER':
        this.handleIncomingOffer(msg);
        return;
      case 'ANSWER':
        this.handleIncomingAnswer(msg);
        return;
      case 'CANDIDATE':
        this.handleIncomingCandidate(msg);
        return;
      case 'ERROR':
        console.error('[Signaling] server error', msg.code);
        return;
    }
  }

  private connectToPeer(peerId: string): void {
    if (!peerId || peerId === this.myId) return;

    const existingPeer = this.peerConnections.get(peerId);
    if (existingPeer) {
      if (existingPeer.isConnected) return;
      existingPeer.destroy();
    }

    const peer = this.constructPeer(peerId, crypto.randomUUID());
    peer.isCaller = true;
    this.peerConnections.set(peerId, peer);
    this.setPeerConnectionState(peerId, 'not connected');
    peer.createPeerConnection();
  }

  private handleIncomingOffer(msg: OfferMessage): void {
    const peerId = msg.peer.id;
    const sessionId = msg.sessionId;

    const existingPeer = this.peerConnections.get(peerId);
    if (existingPeer) {
      existingPeer.destroy();
    }

    const peer = this.constructPeer(peerId, sessionId);
    peer.isCaller = false;
    this.peerConnections.set(peerId, peer);
    this.setPeerConnectionState(peerId, 'not connected');
    void peer.HandlerOffer({ type: 'offer', sdp: msg.sdp });
    this.notify();
  }

  private handleIncomingAnswer(msg: AnswerMessage): void {
    const peer = this.peerConnections.get(msg.peer.id);
    if (!peer) return;
    void peer.HandleAnswer({ type: 'answer', sdp: msg.sdp });
  }

  private handleIncomingCandidate(msg: Extract<WsServerMessage, { type: 'CANDIDATE' }>): void {
    const peer = this.peerConnections.get(msg.peer.id);
    if (!peer || !msg.candidate) return;
    void peer.HandleCandidate(msg.candidate);
  }

  private cleanupPeerConnections(): void {
    this.peerConnections.forEach((peer) => peer.destroy());
    this.peerConnections.clear();
  }

  private connectPeer(peerId: string): void {
    if (this.myId && peerId && peerId !== this.myId && this.myId < peerId) {
      this.connectToPeer(peerId);
    }
  }

  private constructPeer(peerId: string, sessionId: string): Peer {
    const peer = new Peer({
      signaling: this.signaling,
      peerId,
      sessionId,
      onConnectionStateChange: (state) => {
        const status: PeerStatus = state === 'connected' ? 'connected' : 'not connected';
        if (this.setPeerConnectionState(peerId, status)) {
          this.notify();
        }
      },
    });
    peer.iceServers = this.signaling.getIceServers();
    return peer;
  }

  private setPeerConnectionState(peerId: string, status: PeerStatus): boolean {
    if (this.peerConnectionState[peerId] === status) return false;

    this.peerConnectionState = {
      ...this.peerConnectionState,
      [peerId]: status,
    };
    return true;
  }

  private removePeerConnectionState(peerId: string): void {
    if (!(peerId in this.peerConnectionState)) return;
    const next = { ...this.peerConnectionState };
    delete next[peerId];
    this.peerConnectionState = next;
  }
}
