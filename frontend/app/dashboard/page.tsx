"use client";

import Link from "next/link";

import { RequireAuth } from "@/components/RequireAuth";

const CARDS = [
  {
    href: "/play/solo",
    title: "ソロプレイ",
    description: "一人でしりとりを続けよう",
  },
  {
    href: "/play/vs-ai",
    title: "AI対戦",
    description: "AIと交互に単語を出し合う",
  },
  {
    href: "/play/free-talk",
    title: "フリートーク",
    description: "AIと文章しりとりで会話する",
  },
];

export default function DashboardPage() {
  return (
    <RequireAuth>
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <h1 className="mb-6 text-2xl font-bold text-zinc-900">
          ダッシュボード
        </h1>
        <div className="grid gap-4 sm:grid-cols-3">
          {CARDS.map((card) => (
            <Link
              key={card.href}
              href={card.href}
              className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm transition hover:shadow-md"
            >
              <h2 className="mb-2 text-lg font-semibold text-zinc-900">
                {card.title}
              </h2>
              <p className="text-sm text-zinc-600">{card.description}</p>
            </Link>
          ))}
        </div>
      </div>
    </RequireAuth>
  );
}
