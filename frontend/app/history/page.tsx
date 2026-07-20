"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { ApiError, listSessions } from "@/lib/api";
import type { Session } from "@/lib/types";

const MODE_LABELS: Record<string, string> = {
  solo: "ソロプレイ",
  vs_ai: "AI対戦",
  free_talk: "フリートーク",
};

const RESULT_LABELS: Record<string, string> = {
  win: "勝ち",
  lose: "負け",
  ended_by_n: "「ん」で終了",
  ended_by_duplicate: "重複により終了",
  ended_by_user: "自分で終了",
};

function HistoryList() {
  const { authFetch } = useAuth();
  const [sessions, setSessions] = useState<Session[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    authFetch((token) => listSessions(token))
      .then((res) => setSessions(res.sessions ?? []))
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "履歴の取得に失敗しました");
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (error) return <p className="text-sm text-red-600">{error}</p>;
  if (sessions === null) return <p className="text-zinc-500">読み込み中...</p>;
  if (sessions.length === 0) {
    return <p className="text-zinc-500">まだプレイ履歴がありません。</p>;
  }

  return (
    <ul className="divide-y divide-zinc-200 rounded-lg border border-zinc-200 bg-white">
      {sessions.map((s) => (
        <li key={s.id}>
          <Link
            href={`/history/${s.id}`}
            className="flex items-center justify-between px-4 py-3 hover:bg-zinc-50"
          >
            <div>
              <p className="font-medium text-zinc-900">
                {MODE_LABELS[s.mode] ?? s.mode}
              </p>
              <p className="text-xs text-zinc-500">
                {new Date(s.started_at).toLocaleString("ja-JP")}
              </p>
            </div>
            <div className="text-sm text-zinc-600">
              {s.status === "in_progress"
                ? "プレイ中"
                : s.result
                  ? (RESULT_LABELS[s.result] ?? s.result)
                  : "終了"}
            </div>
          </Link>
        </li>
      ))}
    </ul>
  );
}

export default function HistoryPage() {
  return (
    <RequireAuth>
      <div className="mx-auto w-full max-w-2xl px-4 py-8">
        <h1 className="mb-6 text-2xl font-bold text-zinc-900">履歴</h1>
        <HistoryList />
      </div>
    </RequireAuth>
  );
}
