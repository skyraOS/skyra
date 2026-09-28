import type { Metadata } from "next";
import "./globals.css";
import { WSBoot } from "@/components/WSBoot";
import { Sky } from "@/components/Sky";

export const metadata: Metadata = {
  title: "Skyra",
  description: "Chat with Codex through the Skyra runtime.",
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
        <Sky />
        <main className="relative h-full">{children}</main>
      </body>
    </html>
  );
}
