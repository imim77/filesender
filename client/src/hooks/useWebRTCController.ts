import { useSyncExternalStore } from 'react';
import type { WebRTCController } from '../services/webrtccontroller';

export function useWebRTCController(controller: WebRTCController) {
  return useSyncExternalStore(
    (onStoreChange) => controller.subscribe(onStoreChange),
    () => controller.getSnapshot(),
    () => controller.getSnapshot(),
  );
}
