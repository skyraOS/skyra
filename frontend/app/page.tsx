"use client";

import { Conversation, Header } from "@/components/Conversation";
import { HistoryOverlay } from "@/components/HistoryOverlay";
import { Trees } from "@/components/Trees";
import { useAppStore } from "@/lib/store";

export default function Page() {
  const view = useAppStore((s) => s.view);
  return (
    <div className="flex h-full flex-col">
      <Header />
      {view === "chat" && <HistoryOverlay />}
      <main className="reading-matte min-h-0 flex-1">
        {view === "chat" ? <Conversation /> : <Trees />}
      </main>
    </div>
  );
}
