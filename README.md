# FileSender

Inspired by Apple's AirDrop, FileSender is a peer-to-peer file sharing application that enables direct file transfers between browsers using WebRTC data channels, with a Go-based WebSocket signaling server.

### Planned Features
- Enhanced security (authentication, encryption, access controls)
- Docker support for easy deployment
- UI overhaul

### Server (Go)
A WebSocket signaling server that facilitates peer discovery and WebRTC handshake exchange (SDP offers/answers and ICE candidates). The server does **not** handle file transfers - all data goes directly between peers via WebRTC.

