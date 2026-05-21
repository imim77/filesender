import laptopDog from '../assets/laptopdog.png';
import mobileDog from '../assets/mobiledog.png';

type PeerClient = {
  id: string;
  alias?: string;
  deviceModel?: string;
  deviceType?: string;
};

type HeroProps = {
  alias?: string;
  peers?: PeerClient[];
  connectionStatus?: 'disconnected' | 'connecting' | 'connected';
  connectionLabel?: (peerId: string) => string;
  onSendFiles?: (peerId: string, files: FileList) => void;
  className?: string;
};

const peerAvatar = (deviceType?: string): string => (deviceType === 'Mobile' ? mobileDog : laptopDog);

const peerName = (peer: PeerClient): string => peer.alias || peer.deviceModel || peer.deviceType || 'Unnamed device';

const fileInputId = (peerId: string): string => `peer-file-${peerId.replace(/[^a-zA-Z0-9_-]/g, '')}`;

export default function Hero({
  alias = 'Darke Some',
  peers = [],
  connectionStatus = 'connecting',
  connectionLabel = () => 'not connected',
  onSendFiles = () => undefined,
  className = '',
}: HeroProps) {
  return (
    <main data-slot="hero" className={`flex flex-1 justify-center bg-background px-4 py-8 ${className}`.trim()}>
      <section className="w-full max-w-3xl rounded border bg-card p-4">
        <p className="text-sm text-muted-foreground">You are @{alias}</p>
        <p className="mt-1 text-sm text-muted-foreground">Connection: {connectionStatus}</p>

        <h2 className="mt-4 text-base font-semibold">Peers ({peers.length})</h2>

        {peers.length === 0 ? (
          <p className="mt-3 text-sm text-muted-foreground">No peers connected yet.</p>
        ) : (
          <div className="mt-3 space-y-2">
            {peers.map((peer) => (
              <article key={peer.id} className="flex items-center gap-3 rounded border p-3">
                <img src={peerAvatar(peer.deviceType)} alt={peerName(peer)} className="size-10 rounded-full" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{peerName(peer)}</p>
                  <p className="truncate text-xs text-muted-foreground">{connectionLabel(peer.id)}</p>
                </div>

                <input
                  id={fileInputId(peer.id)}
                  type="file"
                  multiple
                  className="text-sm"
                  onChange={(event) => {
                    if (!event.target.files || event.target.files.length === 0) return;
                    onSendFiles(peer.id, event.target.files);
                    event.target.value = '';
                  }}
                />
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
