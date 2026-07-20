"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { ApiError, getSessionDetail } from "@/lib/api";
import type { MessageRecord, Session, WordRecord } from "@/lib/types";

function HistoryDetail() {
  const { id } = useParams<{ id: string }>();
  const { authFetch } = useAuth();
  const [session, setSession] = useState<Session | null>(null);
  const [words, setWords] = useState<WordRecord[] | undefined>(undefined);
  const [messages, setMessages] = useState<MessageRecord[] | undefined>(
    undefined,
  );
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    authFetch((token) => getSessionDetail(token, id))
      .then((res) => {
        setSession(res.session);
        setWords(res.words);
        setMessages(res.messages);
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "取得に失敗しました");
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  if (error) return <p className="text-sm text-red-600">{error}</p>;
  if (!session) return <p className="text-zinc-500">読み込み中...</p>;

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-zinc-200 bg-white p-4">
        <p className="text-sm text-zinc-600">
          モード: {session.mode} / ステータス: {session.status}
          {session.result ? ` / 結果: ${session.result}` : ""}
        </p>
        <p className="text-sm text-zinc-600">ターン数: {session.turn_count}</p>
      </div>

      {words && (
        <ol className="space-y-2">
          {words.map((w) => (
            <li
              key={w.seq_no}
              className={`rounded border p-2 text-sm ${
                w.speaker === "ai"
                  ? "border-blue-200 bg-blue-50"
                  : "border-zinc-200 bg-white"
              }`}
            >
              {w.seq_no}. [{w.speaker === "ai" ? "AI" : "自分"}] {w.word}(
              {w.reading})
            </li>
          ))}
        </ol>
      )}

      {messages && (
        <ol className="space-y-2">
          {messages.map((m) => (
            <li
              key={m.seq_no}
              className={`rounded border p-3 text-sm ${
                m.speaker === "ai"
                  ? "border-blue-200 bg-blue-50"
                  : "border-zinc-200 bg-white"
              }`}
            >
              <span className="font-medium">
                {m.speaker === "ai" ? "AI" : "自分"}:
              </span>{" "}
              {m.content}
            </li>
          ))}
        </ol>
      )}

      <Link href="/history" className="text-sm text-zinc-600 hover:underline">
        履歴一覧に戻る
      </Link>
    </div>
  );
}

export default function HistoryDetailPage() {
  return (
    <RequireAuth>
      <div className="mx-auto w-full max-w-2xl px-4 py-8">
        <h1 className="mb-6 text-2xl font-bold text-zinc-900">履歴詳細</h1>
        <HistoryDetail />
      </div>
    </RequireAuth>
  );
}
