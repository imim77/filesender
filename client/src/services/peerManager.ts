import type {
    ClientInfo,
    OfferMessage,
    SignalingConnection,
    WsServerMessage,
} from './signaling';
import { Peer } from './webrtc';

export class PeerManager {
    signaling: SignalingConnection;
    peersBySessionId: Map<string, Peer> = new Map();
    pendingCandidatesBySessionId: Map<string, RTCIceCandidateInit[]> = new Map();
    knownPeersById: Map<string, ClientInfo> = new Map();
    me: ClientInfo | null = null;

    _onPeerCreated?: (peer: Peer) => void;
    _onPeerRemoved?: (peer: Peer) => void;
    _onError?: (error: unknown) => void;

    constructor(opts: {
        signaling: SignalingConnection;
        onPeerCreated?: (peer: Peer) => void;
        onPeerRemoved?: (peer: Peer) => void;
        onError?: (error: unknown) => void;
    }) {
        this.signaling = opts.signaling;
        this._onPeerCreated = opts.onPeerCreated;
        this._onPeerRemoved = opts.onPeerRemoved;
        this._onError = opts.onError;
    }

    startSession(peerId: string, sessionId: string = crypto.randomUUID()): Peer {
        const existing = this.peersBySessionId.get(sessionId);
        if (existing) {
            return existing;
        }

        console.log('[PeerManager] start session', { sessionId, peerId });

        const peer = new Peer({
            signaling: this.signaling,
            peerId,
            sessionId,
        });

        peer.isCaller = true;
        peer.iceServers = this.signaling.getIceServers();
        peer.createPeerConnection();

        this.peersBySessionId.set(sessionId, peer);
        this._onPeerCreated?.(peer);

        return peer;
    }

    async handleMessage(msg: WsServerMessage): Promise<void> {
        console.log('[PeerManager] handle message', msg.type);
        switch (msg.type) {
            case 'HELLO':
                this.me = msg.client;
                this._setKnownPeers(msg.peers);
                this._refreshIceServers();
                for (const peer of this.knownPeersById.values()) {
                    this._startSessionIfNeeded(peer.id, true);
                }
                return;

            case 'JOIN':
                this._upsertKnownPeer(msg.peer);
                this._startSessionIfNeeded(msg.peer.id, true);
                return;

            case 'UPDATE':
                this._upsertKnownPeer(msg.peer);
                return;

            case 'OFFER':
                await this._handleOffer(msg);
                return;

            case 'ANSWER':
                await this._handleAnswer(msg.sessionId, msg.sdp);
                return;

            case 'CANDIDATE':
                await this._handleCandidate(msg.sessionId, msg.candidate);
                return;

            case 'LEFT':
                this._removeKnownPeer(msg.peerId);
                this.removePeerByPeerId(msg.peerId);
                return;

            default:
                return;
        }
    }

    removePeerByPeerId(peerId: string): void {
        for (const [sessionId, peer] of this.peersBySessionId.entries()) {
            if (peer.peerId !== peerId) continue;
            peer.destroy();
            this.peersBySessionId.delete(sessionId);
            this.pendingCandidatesBySessionId.delete(sessionId);
            this._onPeerRemoved?.(peer);
        }
    }

    getConnectedPeers(): Peer[] {
        return Array.from(this.peersBySessionId.values()).filter((peer) => peer.dc?.readyState === 'open');
    }

    getConnectedPeerCount(): number {
        return this.getConnectedPeers().length;
    }

    getSelf(): ClientInfo | null {
        return this.me;
    }

    getPeers(): ClientInfo[] {
        return Array.from(this.knownPeersById.values());
    }

    sendFilesToConnectedPeers(files: FileList | File[]): { peers: number; files: number } {
        const list = Array.isArray(files) ? files : Array.from(files);
        if (list.length === 0) {
            return { peers: 0, files: 0 };
        }

        const peers = this.getConnectedPeers();
        for (const peer of peers) {
            peer.sendFiles(list);
        }

        return {
            peers: peers.length,
            files: list.length,
        };
    }

    destroy(): void {
        for (const peer of this.peersBySessionId.values()) {
            peer.destroy();
            this._onPeerRemoved?.(peer);
        }
        this.peersBySessionId.clear();
        this.pendingCandidatesBySessionId.clear();
        this.knownPeersById.clear();
        this.me = null;
    }

    private _startSessionIfNeeded(peerId: string, isAutomatic = false): void {
        if (isAutomatic && !this._shouldInitiate(peerId)) {
            console.log('[AUTO CONNECT] skipping (wait for remote offer):', peerId);
            return;
        }

        if (this._hasSessionForPeer(peerId)) {
            console.log('[CONNECT] session already exists:', peerId);
            return;
        }

        console.log(isAutomatic ? '[AUTO CONNECT] starting session to:' : '[CONNECT] starting session to:', peerId);
        this.startSession(peerId);
    }

    private _shouldInitiate(peerId: string): boolean {
        if (!this.me) return false;
        return this.me.id.localeCompare(peerId) < 0;
    }

    private _hasSessionForPeer(peerId: string): boolean {
        for (const peer of this.peersBySessionId.values()) {
            if (peer.peerId === peerId) return true;
        }
        return false;
    }

    private _setKnownPeers(peers: ClientInfo[]): void {
        this.knownPeersById.clear();
        for (const peer of peers) {
            this._upsertKnownPeer(peer);
        }
    }

    private _upsertKnownPeer(peer: ClientInfo): void {
        if (this.me && peer.id === this.me.id) return;
        this.knownPeersById.set(peer.id, peer);
    }

    private _removeKnownPeer(peerId: string): void {
        this.knownPeersById.delete(peerId);
    }

    private _refreshIceServers(): void {
        const iceServers = this.signaling.getIceServers();
        for (const peer of this.peersBySessionId.values()) {
            if (peer.pc) continue;
            peer.iceServers = iceServers;
        }
    }

    private async _handleOffer(msg: OfferMessage): Promise<void> {
        let peer = this.peersBySessionId.get(msg.sessionId);
        if (!peer) {
            peer = new Peer({
                signaling: this.signaling,
                peerId: msg.peer.id,
                sessionId: msg.sessionId,
            });
            peer.isCaller = false;
            peer.iceServers = this.signaling.getIceServers();
            this.peersBySessionId.set(msg.sessionId, peer);
            this._onPeerCreated?.(peer);
        }

        try {
            await peer.HandlerOffer({ type: 'offer', sdp: msg.sdp });
            await this._flushPendingCandidates(msg.sessionId, peer);
        } catch (error) {
            this._onError?.(error);
        }
    }

    private async _handleAnswer(sessionId: string, sdp: string): Promise<void> {
        const peer = this.peersBySessionId.get(sessionId);
        if (!peer) return;

        try {
            await peer.HandleAnswer({ type: 'answer', sdp });
        } catch (error) {
            this._onError?.(error);
        }
    }

    private async _handleCandidate(sessionId: string, candidate: RTCIceCandidateInit | null): Promise<void> {
        if (!candidate) return;

        const peer = this.peersBySessionId.get(sessionId);
        if (!peer) {
            const pending = this.pendingCandidatesBySessionId.get(sessionId) ?? [];
            pending.push(candidate);
            this.pendingCandidatesBySessionId.set(sessionId, pending);
            return;
        }

        try {
            await peer.HandleCandidate(candidate);
        } catch (error) {
            this._onError?.(error);
        }
    }

    private async _flushPendingCandidates(sessionId: string, peer: Peer): Promise<void> {
        const pending = this.pendingCandidatesBySessionId.get(sessionId);
        if (!pending || pending.length === 0) return;

        this.pendingCandidatesBySessionId.delete(sessionId);

        for (const candidate of pending) {
            await peer.HandleCandidate(candidate);
        }
    }
}
