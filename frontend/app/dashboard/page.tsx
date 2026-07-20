"use client";

import { Bot, MessageCircle, User } from "lucide-react";
import Link from "next/link";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";

const CARDS = [
  {
    href: "/play/solo",
    title: "ソロプレイ",
    description: "一人でしりとりを続けよう",
    icon: User,
    iconBg: "bg-slate-900",
  },
  {
    href: "/play/vs-ai",
    title: "AI対戦",
    description: "AIと交互に単語を出し合う",
    icon: Bot,
    iconBg: "bg-blue-600",
  },
  {
    href: "/play/free-talk",
    title: "フリートーク",
    description: "AIと文章しりとりで会話する",
    icon: MessageCircle,
    iconBg: "bg-rose-500",
  },
];

function DashboardContent() {
  const { user } = useAuth();

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <main className="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="mb-10 rounded-2xl bg-white p-8 shadow-sm ring-1 ring-black/5">
          <h1 className="text-2xl font-bold tracking-tight">
            おかえりなさい、{user?.display_name}さん
          </h1>
          <p className="mt-2 text-muted-foreground">
            モードを選んでプレイを始めましょう
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {CARDS.map(({ href, title, description, icon: Icon, iconBg }) => (
            <Link key={href} href={href}>
              <div className="flex h-full flex-col gap-4 rounded-2xl bg-white p-6 shadow-sm ring-1 ring-black/5 transition-shadow hover:shadow-md">
                <div
                  className={`flex size-12 items-center justify-center rounded-full text-white ${iconBg}`}
                >
                  <Icon className="size-6" />
                </div>
                <div>
                  <h2 className="font-semibold">{title}</h2>
                  <p className="mt-1 text-sm text-muted-foreground">{description}</p>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </main>
    </div>
  );
}

export default function DashboardPage() {
  return (
    <RequireAuth>
      <DashboardContent />
    </RequireAuth>
  );
}
