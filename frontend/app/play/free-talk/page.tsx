"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import {
  ApiError,
  createSession,
  endSession,
  submitMessage,
} from "@/lib/api";

type ChatMessage = { speaker: "user" | "ai"; content: string };

function FreeTalkPlay() {
  const { authFetch } = useAuth();
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [finished, setFinished] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    authFetch((token) => createSession(token, "free_talk"))
      .then((s) => setSessionId(s.id))
      .catch((err) => {
        setErrorMessage(err instanceof ApiError ? err.message : "開始に失敗しました");
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSubmit() {
    if (!sessionId || !input.trim() || submitting || finished) return;
    const content = input;
    setInput("");
    setSubmitting(true);
    setErrorMessage(null);
    setMessages((m) => [...m, { speaker: "user", content }]);
    try {
      const res = await authFetch((token) => submitMessage(token, sessionId, content));
      if (res.ai_message) {
        setMessages((m) => [
          ...m,
          { speaker: "ai", content: res.ai_message!.content },
        ]);
      }
      if (res.status === "finished") {
        setFinished(res.result ?? "ended");
      }
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "通信エラーが発生しました");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleEnd() {
    if (!sessionId) return;
    try {
      await authFetch((token) => endSession(token, sessionId));
      setFinished("ended_by_user");
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "終了に失敗しました");
    }
  }

  if (loading) return <p className="text-zinc-500">準備中...</p>;

  return (
    <div className="flex flex-1 flex-col">
      <div className="flex-1 space-y-3 overflow-y-auto rounded-lg border border-zinc-200 bg-white p-4">
        {messages.length === 0 && (
          <p className="text-center text-sm text-zinc-400">
            最初のメッセージを送ってみましょう(ひらがな・カタカナ・英字で終わる文にしてください)
          </p>
        )}
        {messages.map((m, i) => (
          <div
            key={i}
            className={`flex ${m.speaker === "user" ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`max-w-[75%] rounded-lg px-3 py-2 text-sm ${
                m.speaker === "user"
                  ? "bg-zinc-900 text-white"
                  : "bg-blue-50 text-zinc-900"
              }`}
            >
              {m.content}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {finished ? (
        <div className="mt-4 space-y-3 text-center">
          <p className="font-medium text-zinc-900">会話が終了しました</p>
          <div className="flex justify-center gap-4">
            <Link
              href="/play/free-talk"
              className="rounded bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700"
            >
              もう一度話す
            </Link>
            <Link
              href="/dashboard"
              className="rounded border border-zinc-300 px-4 py-2 text-zinc-700 hover:bg-zinc-100"
            >
              ダッシュボードへ
            </Link>
          </div>
        </div>
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="mt-4 flex gap-2"
        >
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="メッセージを入力"
            className="flex-1 rounded border border-zinc-300 px-3 py-2 focus:border-zinc-500 focus:outline-none"
            autoFocus
          />
          <button
            type="submit"
            disabled={submitting}
            className="rounded bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700 disabled:opacity-50"
          >
            送信
          </button>
          <button
            type="button"
            onClick={handleEnd}
            className="rounded border border-zinc-300 px-3 py-2 text-sm text-zinc-600 hover:bg-zinc-100"
          >
            終了する
          </button>
        </form>
      )}

      {errorMessage && (
        <p className="mt-2 text-sm text-red-600">{errorMessage}</p>
      )}
    </div>
  );
}

export default function FreeTalkPlayPage() {
  return (
    <RequireAuth>
      <div className="mx-auto flex w-full max-w-xl flex-1 flex-col px-4 py-8">
        <h1 className="mb-6 text-center text-2xl font-bold text-zinc-900">
          フリートーク
        </h1>
        <FreeTalkPlay />
      </div>
    </RequireAuth>
  );
}
