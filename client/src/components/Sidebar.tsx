import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

type SidebarProps = {
  alias: string;
  browserPeer: string;
  connectionStatus: string;
  peerCount: number;
};

export default function Sidebar({ alias, browserPeer, connectionStatus, peerCount }: SidebarProps) {
  return (
    <aside className="flex h-full min-h-0 flex-col gap-6 bg-sidebar p-6">
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <p className="text-sm uppercase tracking-wide text-muted-foreground">
            You are visible as
          </p>
          <h2 className="text-2xl font-semibold text-foreground">
            {alias || "Anonymous"}
          </h2>
        </div>
        <br />
        <div className="flex flex-col gap-3">
          <p className="text-sm leading-relaxed text-muted-foreground">
            Make sure both devices are unlocked, close together, and have
            Bluetooth turned on. Devices you're sharing with need Quick Share
            turned on and visible to you.{" "}
          </p>
        </div>
      </div> 
    </aside>
  );
}
