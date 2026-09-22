import type { Metadata } from "next";
import "./globals.css";
import { AppNav } from "@/components/AppNav";
import { WSBoot } from "@/components/WSBoot";

export const metadata: Metadata = {
  title: "AI Beings",
  description: "Render of a Logos/Relations runtime.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="h-screen overflow-hidden">
        <WSBoot />
        <div className="flex h-full flex-col">
          <AppNav />
          <main className="flex-1 min-h-0">{children}</main>
        </div>
      </body>
    </html>
  );
}
