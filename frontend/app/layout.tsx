import type { Metadata } from "next";
import "./globals.css";

import { AuthProvider } from "@/components/AuthProvider";
import { Header } from "@/components/Header";
import { Geist } from "next/font/google";
import { cn } from "@/lib/utils";

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

export const metadata: Metadata = {
  title: "しりとりアプリ",
  description: "ソロプレイ・AI対戦・文章しりとりで遊べるしりとりWebアプリ",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja" className={cn("h-full", "font-sans", geist.variable)}>
      <body className="min-h-full flex flex-col bg-background font-sans antialiased">
        <AuthProvider>
          <Header />
          <div className="flex flex-1 flex-col">{children}</div>
        </AuthProvider>
      </body>
    </html>
  );
}
