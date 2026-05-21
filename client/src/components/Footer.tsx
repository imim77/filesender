import { useEffect, useMemo, useState } from 'react';
import { getBrowser } from '../utilis/uaNames';

export default function Footer() {
  const browser = useMemo(() => getBrowser(navigator.userAgent), []);
  const fullCommitHash = import.meta.env.VITE_GIT_COMMIT || import.meta.env.PUBLIC_GIT_COMMIT || 'local-dev';
  const shortCommitHash = fullCommitHash.slice(0, 7);
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  const timeText = now.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });

  return (
    <footer className="w-full bg-background/80 backdrop-blur-sm">
      <div className="mx-auto grid w-full max-w-screen-xl grid-cols-3 gap-4 px-4 py-3 text-sm">
        <div className="flex flex-col items-center gap-0.5">
          <span className="text-xs text-muted-foreground">Browser:</span>
          <span className="text-foreground">{browser}</span>
        </div>
        <div className="flex flex-col items-center gap-0.5">
          <span className="text-xs text-muted-foreground">Commit:</span>
          <span className="text-foreground">{shortCommitHash}</span>
        </div>
        <div className="flex flex-col items-center gap-0.5">
          <span className="text-xs text-muted-foreground">Time:</span>
          <span className="text-foreground">{timeText}</span>
        </div>
      </div>
    </footer>
  );
}
