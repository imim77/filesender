
export class SignalingConnection extends EventTarget{
    socket: WebSocket|null = null;
    url: string;
    info: ClientInfoWithoutId|null;
    private _iceServers: IceServerInfo[] = [];


    constructor({info}: {info?: ClientInfoWithoutId}){
        super()
        this.url = this.endpoint
        this.info = info ?? null;
    }

    connect(){
        const ws = new WebSocket(this.url)
        ws.onopen = ()=>{
            console.log('WS: signaling connection established');
            this.dispatchEvent(new CustomEvent('open'));

            if (this.info) {
                this.send({
                    type: 'UPDATE',
                    info: this.info,
                });
            }
        }

        ws.onmessage = (event)=>{ 
            let msg: WsServerMessage;
            try{
                msg = JSON.parse(event.data) as WsServerMessage;
            }catch(error){
                console.error('Failed to parse signaling message', error);
                return;
            }
            return this.onMessage(msg);

        }

        ws.onerror = (error)=>{
            console.error('WS: signaling error', error);
            this.dispatchEvent(new CustomEvent('error', { detail: error }));
        }

        this.socket = ws;
    }

    onMessage(msg: WsServerMessage){
        if(msg.type === 'HELLO' && msg.iceServers){this._iceServers = msg.iceServers};
        if(msg.type === 'ANSWER'){}
    }

    get endpoint(){
        const protocol = location.protocol.startsWith('https') ? 'wss' : 'ws';
        const port = '9000';
        return `${protocol}://${location.hostname}:${port}/ws`;
    }

    send(msg: WsClientMessage){
        console.log("WS sending to the server: ", msg)
        this.socket?.send(JSON.stringify(msg))
    }



}




export interface ClientInfoWithoutId {
    alias: string;
    deviceModel?: string;
    deviceType?: string;
    token?: string;
}

export interface ClientInfo extends ClientInfoWithoutId {
    id: string;
}

export type WsClientMessage =
    | { type: 'OFFER'; sessionId: string; target: string; sdp: string }
    | { type: 'ANSWER'; sessionId: string; target: string; sdp: string }
    | { type: 'CANDIDATE'; sessionId: string; target: string; candidate: RTCIceCandidateInit | null }
    | { type: 'UPDATE'; info: ClientInfoWithoutId };

export type WsServerMessage =
    | { type: 'HELLO'; client: ClientInfo; peers: ClientInfo[]; iceServers?: IceServerInfo[] }
    | { type: 'JOIN'; peer: ClientInfo }
    | { type: 'UPDATE'; peer: ClientInfo }
    | { type: 'LEFT'; peerId: string }
    | OfferMessage
    | AnswerMessage
    | { type: 'CANDIDATE'; peer: ClientInfo; sessionId: string; candidate: RTCIceCandidateInit | null }
    | { type: 'ERROR'; code: number };

export interface OfferMessage {
    type: 'OFFER';
    peer: ClientInfo;
    sessionId: string;
    sdp: string;
}

export interface AnswerMessage {
    type: 'ANSWER';
    peer: ClientInfo;
    sessionId: string;
    sdp: string;
}

export interface IceServerInfo {
    urls: string[];
    username?: string;
    credential?: string;
}