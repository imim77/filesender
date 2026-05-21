import { useEffect, useMemo } from 'react';
import githubLogo from './assets/githublogo.svg';
import Footer from './components/Footer';
import Hero from './components/Hero';
import Navigation from './components/Navigation';
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
    <div className="flex min-h-screen flex-col">
      <Navigation logoSrc={githubLogo} />
      <Hero
        alias={state.myName || localAlias}
        peers={state.peers}
        connectionStatus={state.connectionStatus}
        connectionLabel={(peerId) => controller.connectionLabel(peerId)}
        onSendFiles={(peerId, files) => controller.sendFiles(peerId, files)}
      />
      <div className="mt-auto">
        <Footer />
      </div>
    </div>
  );
}
