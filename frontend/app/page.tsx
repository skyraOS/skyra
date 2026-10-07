import { Conversation, Header } from "@/components/Conversation";
import { HistoryOverlay } from "@/components/HistoryOverlay";

export default function Page() {
  return (
    <div className="flex h-full flex-col">
      <Header />
      <HistoryOverlay />
      <main className="reading-matte min-h-0 flex-1">
        <Conversation />
      </main>
    </div>
  );
}
