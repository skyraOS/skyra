import { Conversation } from "@/components/Conversation";
import { Sidebar } from "@/components/Sidebar";

export default function Page() {
  return (
    <div className="flex h-full">
      <Sidebar />
      <div className="h-full flex-1">
        <Conversation />
      </div>
    </div>
  );
}
