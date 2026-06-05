import type { ClientInfo } from '../services/signaling';
import DeviceCard from './DeviceCard';
import { Badge } from './ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from './ui/card';

type DeviceListProps = {
  peers: ClientInfo[];
  peerConnectionState: Record<string, string>;
  onSendFiles: (peerId: string, files: FileList) => void;
};

export default function DeviceList({ peers, peerConnectionState, onSendFiles }: DeviceListProps) {
  return (
    <div className="flex min-h-full flex-col gap-4 p-4 md:p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">All peers</h2>
          <p className="text-sm text-muted-foreground">Nearby browser sessions available for transfer.</p>
        </div>
        <Badge variant="outline">{peers.length} total</Badge>
      </div>

      {peers.length === 0 ? (
        <Card className="flex min-h-56 items-center justify-center">
          <CardHeader className="items-center text-center">
            <CardTitle>No peers found</CardTitle>
            <CardDescription>
              Open Filesender in another browser on the same network to start sharing.
            </CardDescription>
          </CardHeader>
        </Card>
      ) : (
        <Card>
          {peers.map((peer) => (
            <DeviceCard
              key={peer.id}
              peer={peer}
              connectionStatus={peerConnectionState[peer.id] ?? 'not connected'}
              onSendFiles={onSendFiles}
            />
          ))} 
        </Card>
      )}
    </div>
  );
}
