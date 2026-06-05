import { useEffect, useMemo } from 'react';
import githubLogo from './assets/githublogo.svg';
import Navigation from './components/Navigation';
import QuickShareLayout from './components/QuickShareLayout';
import { useWebRTCController } from './hooks/useWebRTCController';
import { WebRTCController } from './services/webrtccontroller';
import { generateName, getAgentInfo } from './utilis/uaNames';

export default function App() {
  const localAlias = useMemo(() => generateName(), []);
  const localDevice = useMemo(() => getAgentInfo(navigator.userAgent), []);
  const controller = useMemo(() => new WebRTCController(localAlias, localDevice), [localAlias, localDevice]);
  const state = useWebRTCController(controller);

  useEffect(() => {
    return () => controller.destroy();
  }, [controller]);

  return (
    <div className="flex h-dvh min-h-0 flex-col bg-muted/30">
      <Navigation logoSrc={githubLogo} />
      <div className="flex flex-1 overflow-hidden p-4 md:p-6">
        <div className="flex w-full flex-col overflow-hidden rounded-xl border bg-background shadow-sm">
          <QuickShareLayout
            alias={state.myName || localAlias}
            browserPeer={localDevice}
            peers={state.peers}
            connectionStatus={state.connectionStatus}
            peerConnectionState={Object.fromEntries(
              state.peers.map((peer) => [peer.id, controller.connectionLabel(peer.id)])
            )}
            onSendFiles={(peerId, files) => controller.sendFiles(peerId, files)}
          />
        </div>
      </div>
    </div>
  );
}
