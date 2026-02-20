
import type { IceServerInfo, SignalingConnection } from "./signaling";

export class Peer{
    pc: RTCPeerConnection|null = null;
    dc: RTCDataChannel|null = null;
    signaling: SignalingConnection;
    peerId: string;
    sessionId: string;
    isConnected: boolean = false;
    iceServers: IceServerInfo[] = [];

    constructor({signaling, peerId, sessionId
    }: {
        signaling: SignalingConnection;
        peerId: string;
        sessionId: string;
    }){
        this.signaling = signaling;
        this.peerId = peerId;
        this.sessionId = sessionId;
    }

    createPeerConnection(){
        if(this.pc) return;
        const config: RTCConfiguration = {
            iceServers: this.iceServers.map(server => ({
                urls: server.urls,
                username: server.username,
                credential: server.credential,
            })),
        };

        this.pc = new RTCPeerConnection(config);
        this.pc.onicecandidate = (event) => {
            if(event.candidate){
                this.signaling.send({
                    type:'CANDIDATE',
                    sessionId: this.sessionId,
                    target: this.peerId,
                    candidate: event.candidate,
                })
            }
        }
        this.pc.onconnectionstatechange = () => {
            const state = this.pc?.connectionState;
            if (state === 'connected') {
                this.isConnected = true; 
            } else if (state === 'disconnected' || state === 'failed' || state === 'closed') {
                this.isConnected = false; 
            }
        }

        this.pc.ondatachannel = (ev) => {
            
        }


    }

    setupDataChannel(dc: RTCDataChannel){
        this.dc = dc;
        dc.binaryType = 'arraybuffer';
        dc.onopen = () => {
            console.log("Data channel opeend");
        }
    }

    createDataChannel(){
        if(!this.pc) return;
        const dc = this.pc.createDataChannel('data', {ordered: true})
        

    }
}