

import type { IceServerInfo, SignalingConnection } from "./signaling";

export class Peer{
    pc: RTCPeerConnection|null = null;
    dc: RTCDataChannel|null = null;
    signaling: SignalingConnection;
    peerId: string;
    sessionId: string;
    isConnected: boolean = false;
    isCaller: boolean = false;
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
            this.setupDataChannel(ev.channel);
        }

        this.createDataChannel();

        if(this.isCaller){
            this.createOffer();
        }

    }

    setupDataChannel(dc: RTCDataChannel){
        this.dc = dc;
        dc.binaryType = 'arraybuffer';
        dc.onopen = () => {
            console.log("Data channel opeend");
        }
        dc.onmessage = () => {
            console.log("Process message on message");
        }

        dc.onclose = ()=>{
            console.log('Data channel closed');
        }

        dc.onerror = (error) => {
            console.error('Data channel error: ', error)
        }
    }

    createDataChannel(){
        if(!this.pc) return;
        const dc = this.pc.createDataChannel('data', {ordered: true})
        this.setupDataChannel(dc)
        

    }

    private async createOffer(){
        if(!this.pc) return;
        try{
            const offer = await this.pc.createOffer();
            await this.pc.setLocalDescription(offer);
            this.signaling.send({
                type: "OFFER",
                sessionId: this.sessionId,
                target: this.peerId,
                sdp: offer.sdp!,
            })
        }catch(error){  
            console.error('Failed to create offer: ', error);
        }
    }

    async HandlerOffer(offer: RTCSessionDescriptionInit){
        if(!this.pc){
            this.createPeerConnection()
        }
        if(!this.pc) return;

        try{
            await this.pc.setRemoteDescription(offer);
            const answer = await this.pc.createAnswer();
            await this.pc.setLocalDescription(answer);

            this.signaling.send({
                type: "ANSWER",
                sessionId: this.sessionId,
                target: this.peerId,
                sdp: answer.sdp!,
            })
        }catch(error){
            console.error('Failed to handle offer:', error);
        }
        
    }

    async HandleAnswer(answer: RTCSessionDescriptionInit){
        if(!this.pc) return;

        try{
            await this.pc.setRemoteDescription(answer);
        }catch(error){
            console.error('Failed to handle answer:', error);
        }
    }

    async HandleCandidate(candidate: RTCIceCandidateInit | RTCIceCandidate){
        if(!this.pc) return;

        try{
            const iceCandidate = candidate instanceof RTCIceCandidate ? candidate : new RTCIceCandidate(candidate);
            this.pc.addIceCandidate(iceCandidate)
        }catch(error){
            console.error('Failed to add ICE candidate', error)
        }
        
            
    }

    destroy(){
        if (this.dc) {
            this.dc.onopen = null;
            this.dc.onmessage = null;
            this.dc.onclose = null;
            this.dc.onerror = null;
            this.dc.close();
            this.dc = null;
        }

        if (this.pc) {
            this.pc.onicecandidate = null;
            this.pc.onconnectionstatechange = null;
            this.pc.ondatachannel = null;
            this.pc.close();
            this.pc = null;
        }

        this.isConnected = false;
    }

}
