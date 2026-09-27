import type { Metadata } from "next";
import "./globals.css";
import { WSBoot } from "@/components/WSBoot";

export const metadata: Metadata = {
  title: "Skyra",
  description: "Browser client for a DNA-19 user-executable interaction.",
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
        <main className="h-full">{children}</main>
      </body>
    </html>
  );
}
