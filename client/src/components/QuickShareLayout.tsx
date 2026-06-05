import type { ClientInfo } from '../services/signaling';
import { useSidebarResize } from '../hooks/useSidebarResize';
import DeviceList from './DeviceList';
import Sidebar from './Sidebar';

type QuickShareLayoutProps = {
  alias: string;
  browserPeer: string;
  connectionStatus: string;
  peers: ClientInfo[];
  peerConnectionState: Record<string, string>;
  onSendFiles: (peerId: string, files: FileList) => void;
};

export default function QuickShareLayout({
  alias,
  browserPeer,
  connectionStatus,
  peers,
  peerConnectionState,
  onSendFiles,
}: QuickShareLayoutProps) {
  const { isCollapsed, isDragging, handleMouseDown, handleDoubleClick, sidebarRef, overlayRef } =
    useSidebarResize();

  return (
    <main className="flex h-full min-h-0 flex-col">
      {/* Mobile layout */}
      <div className="flex min-h-0 flex-1 flex-col md:hidden">
        <section className="min-h-[18rem] overflow-y-auto rounded-tl-xl border-b bg-muted/20">
          <Sidebar
            alias={alias}
            browserPeer={browserPeer}
            connectionStatus={connectionStatus}
            peerCount={peers.length}
          />
        </section>
        <section className="min-h-0 flex-1 overflow-y-auto rounded-bl-xl">
          <DeviceList
            peers={peers}
            peerConnectionState={peerConnectionState}
            onSendFiles={onSendFiles}
          />
        </section>
      </div>

      {/* Desktop layout with draggable panel */}
      <div className="hidden h-full min-h-0 w-full md:flex">
        {/* Sidebar panel */}
        <div
          ref={sidebarRef}
          className="relative h-full shrink-0 overflow-hidden border-r"
          style={{ width: 300, willChange: 'width' }}
        >
          <div className="h-full min-w-[200px]">
            <Sidebar
              alias={alias}
              browserPeer={browserPeer}
              connectionStatus={connectionStatus}
              peerCount={peers.length}
            />
          </div>
          {/* Gray overlay that fades in as panel narrows */}
          <div
            ref={overlayRef}
            className="pointer-events-none absolute inset-0 bg-background"
            style={{ opacity: 0, willChange: 'opacity' }}
          />
        </div>

        {/* Drag handle */}
        <div
          onMouseDown={handleMouseDown}
          onDoubleClick={handleDoubleClick}
          className={`group relative z-10 flex w-1.5 cursor-col-resize items-center justify-center select-none ${
            isDragging ? 'bg-primary/10' : 'hover:bg-primary/10'
          }`}
        >
          <div className="h-8 w-1 rounded-full bg-border transition-colors group-hover:bg-primary/40 group-active:bg-primary/50" />
        </div>

        {/* Main content */}
        <div className="h-full min-w-0 flex-1 overflow-y-auto">
          <DeviceList
            peers={peers}
            peerConnectionState={peerConnectionState}
            onSendFiles={onSendFiles}
          />
        </div>
      </div>
    </main>
  );
}
