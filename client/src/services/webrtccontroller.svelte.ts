import { generateName, getAgentInfo } from '../utilis/uaNames';
import { type AnswerMessage, type ClientInfo, SignalingConnection, type OfferMessage, type WsServerMessage } from './signaling';
import { Peer } from './webrtc';

export class WebRTCController {
    connectionStatus = $state<'disconnected' | 'connecting' | 'connected'>('connecting');
    myId = $state('');
    myName = $state('');
    signaling: SignalingConnection;
    peers = $state<ClientInfo[]>([]);
    peerConnections = $state<Map<string, Peer>>(new Map());

    constructor(alias: string, deviceModel: string) {
        this.signaling = new SignalingConnection({
            info: {
                alias: alias,
                deviceModel: deviceModel,
            },
            onOpen: () => {
                this.connectionStatus = 'connecting';
            },
            onMessage: (msg) => {
                this.handleSignalingMessage(msg);
            },
            onClose: () => {
                this.connectionStatus = 'disconnected';
                this.cleanupPeerConnections();
                this.peers = [];
            },
            onError: (error) => {
                this.connectionStatus = 'disconnected';
                console.error('[Signaling] connection error', error);
            },
        });
    }

    handleSignalingMessage(msg: WsServerMessage): void {
        switch (msg.type) {
            case 'HELLO':
                this.myId = msg.client.id;
                this.myName = msg.client.alias;
                this.peers = msg.peers.filter((peer) => peer.id !== this.myId);
                this.connectionStatus = 'connected';

                //auto-connect to all the peers
                this.peers.forEach((peer) => {
                    if(!this.peerConnections.has(peer.id)){
                        this.connectToPeer(peer.id);
                    }
                })
                return;
            case 'JOIN':
                if (msg.peer.id !== this.myId && !this.peers.some((peer) => peer.id === msg.peer.id)) {
                    this.peers = [...this.peers, msg.peer];
                    this.connectToPeer(msg.peer.id);
                }
                return;
            case 'UPDATE':
                if (msg.peer.id === this.myId) return;
                this.peers = this.peers.some((peer) => peer.id === msg.peer.id)
                    ? this.peers.map((peer) => (peer.id === msg.peer.id ? msg.peer : peer))
                    : [...this.peers, msg.peer];
                return;
            case 'LEFT': {
                const peer = this.peerConnections.get(msg.peerId);
                if (peer) {
                    peer.destroy();
                    this.peerConnections.delete(msg.peerId);
                }
                this.peers = this.peers.filter((p) => p.id !== msg.peerId);
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

    connectToPeer(peerId: string): void {
        if (!peerId || peerId === this.myId) return;

        const existingPeer = this.peerConnections.get(peerId);
        if (existingPeer) {
            if (existingPeer.isConnected) return;
            existingPeer.destroy();
        }

        const peer = new Peer({
            signaling: this.signaling,
            peerId,
            sessionId: crypto.randomUUID(),
        });

        peer.isCaller = true;
        peer.iceServers = this.signaling.getIceServers();
        this.peerConnections.set(peerId, peer);
        peer.createPeerConnection();
    }

    sendFiles(peerId: string, files: FileList | File[]): void {
        const peer = this.peerConnections.get(peerId);
        if (!peer) return;
        peer.sendFiles(files);
    }

    handleIncomingOffer(msg: OfferMessage): void {
        const peerId = msg.peer.id;
        const sessionId = msg.sessionId;

        const existingPeer = this.peerConnections.get(peerId);
        if (existingPeer) {
            existingPeer.destroy();
        }

        const peer = new Peer({
            signaling: this.signaling,
            peerId,
            sessionId,
        });

        peer.isCaller = false;
        peer.iceServers = this.signaling.getIceServers();
        this.peerConnections.set(peerId, peer);
        void peer.HandlerOffer({ type: 'offer', sdp: msg.sdp });
    }

    handleIncomingAnswer(msg: AnswerMessage): void {
        const peer = this.peerConnections.get(msg.peer.id);
        if (!peer) return;
        void peer.HandleAnswer({ type: 'answer', sdp: msg.sdp });
    }

    handleIncomingCandidate(msg: Extract<WsServerMessage, { type: 'CANDIDATE' }>): void {
        const peer = this.peerConnections.get(msg.peer.id);
        if (!peer || !msg.candidate) return;
        void peer.HandleCandidate(msg.candidate);
    }

    destroy(): void {
        this.cleanupPeerConnections();
        this.signaling.destroy();
        this.peers = [];
        this.connectionStatus = 'disconnected';
    }

    private cleanupPeerConnections(): void {
        this.peerConnections.forEach((peer) => peer.destroy());
        this.peerConnections.clear();
    }


}
