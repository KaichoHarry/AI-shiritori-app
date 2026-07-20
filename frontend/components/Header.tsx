"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/AuthProvider";

export function Header() {
  const { user, logout } = useAuth();
  const router = useRouter();

  if (!user) return null;

  return (
    <header className="border-b border-zinc-200 bg-white">
      <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-3">
        <Link href="/dashboard" className="text-lg font-bold text-zinc-900">
          しりとりアプリ
        </Link>
        <nav className="flex items-center gap-4 text-sm text-zinc-600">
          <Link href="/dashboard" className="hover:text-zinc-900">
            ダッシュボード
          </Link>
          <Link href="/settings" className="hover:text-zinc-900">
            設定
          </Link>
          <Link href="/history" className="hover:text-zinc-900">
            履歴
          </Link>
          <span className="text-zinc-400">|</span>
          <span>{user.display_name}さん</span>
          <button
            type="button"
            onClick={() => {
              logout();
              router.replace("/login");
            }}
            className="rounded border border-zinc-300 px-2 py-1 hover:bg-zinc-100"
          >
            ログアウト
          </button>
        </nav>
      </div>
    </header>
  );
}
