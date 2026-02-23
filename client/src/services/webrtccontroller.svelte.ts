import { SignalingConnection, type WsServerMessage } from "./signaling";

export class WebRTCController{
    connectionStatus = $state<'disconnected'| 'connecting'| 'connected'>('disconnected');
    myId = $state('')
    signaling: SignalingConnection;

    constructor(alias: string){
        this.signaling = new SignalingConnection({
            info: {alias},
            
        })
    }

        
    handleSignalingMessage(msg: WsServerMessage){
        switch(msg.type){
            case 'HELLO':
                this.myId = msg.client.id;
                return;
            case 'JOIN':
                console.log('Peer joined', msg.peer);
                return;
            case 'OFFER':
                console.log('Recieved answer', msg);
                return;
            case 'ANSWER':
                console.log('Recieved answer', msg);
            
        }
    }
}