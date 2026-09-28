import { Conversation, Header } from "@/components/Conversation";
import { Sidebar } from "@/components/Sidebar";

export default function Page() {
  return (
    <div className="flex h-full flex-col">
      <Header />
      <div className="flex min-h-0 flex-1">
        <Sidebar />
        <div className="h-full min-w-0 flex-1">
          <Conversation />
        </div>
      </div>
    </div>
  );
}
