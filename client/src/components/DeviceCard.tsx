import { useRef } from 'react';
import { Monitor, Send, Smartphone } from 'lucide-react';
import type { ClientInfo } from '../services/signaling';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Card, CardContent } from './ui/card';

type DeviceCardProps = {
  peer: ClientInfo;
  connectionStatus: string;
  onSendFiles: (peerId: string, files: FileList) => void;
};

function isMobile(peer: ClientInfo): boolean {
  const model = (peer.deviceModel || '').toLowerCase();
  return model.includes('mobile') || model.includes('ios') || model.includes('android');
}

function peerName(peer: ClientInfo): string {
  return peer.alias || peer.deviceModel || peer.deviceType || 'Unnamed device';
}

export default function DeviceCard({ peer, connectionStatus, onSendFiles }: DeviceCardProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const isConnected = connectionStatus === 'connected';

  return (
    <Card className={isConnected ? 'bg-accent' : undefined}>
      <CardContent className="flex items-center gap-4 p-4">
        <div className="flex size-11 shrink-0 items-center justify-center rounded-md border bg-background">
          {isMobile(peer) ? <Smartphone /> : <Monitor />}
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{peerName(peer)}</p>
          <p className="truncate text-sm text-muted-foreground">
            {peer.deviceModel || peer.deviceType || 'Unknown device'}
          </p>
        </div>
        <Badge variant={isConnected ? 'default' : 'secondary'}>
          {connectionStatus}
        </Badge>
        <Button type="button" variant="outline" onClick={() => inputRef.current?.click()}>
          <Send data-icon="inline-start" />
          Send
        </Button>
      <input
        ref={inputRef}
        type="file"
        multiple
        className="hidden"
        onChange={(e) => {
          if (e.target.files?.length) {
            onSendFiles(peer.id, e.target.files);
            e.target.value = '';
          }
        }}
      />
      </CardContent>
    </Card>
  );
}
