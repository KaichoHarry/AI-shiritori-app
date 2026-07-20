"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { ApiError, createSession, submitWord } from "@/lib/api";
import type { WordView } from "@/lib/types";

function SoloPlay() {
  const { authFetch } = useAuth();
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [lastWord, setLastWord] = useState<WordView | null>(null);
  const [history, setHistory] = useState<WordView[]>([]);
  const [input, setInput] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [gameOver, setGameOver] = useState<{ reason?: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    authFetch((token) => createSession(token, "solo"))
      .then((s) => setSessionId(s.id))
      .catch((err) => {
        setErrorMessage(err instanceof ApiError ? err.message : "開始に失敗しました");
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function handleSubmit() {
    if (!sessionId || !input.trim() || submitting) return;
    setSubmitting(true);
    setErrorMessage(null);
    try {
      const res = await authFetch((token) => submitWord(token, sessionId, input));
      if (res.accepted && res.player_word) {
        setLastWord(res.player_word);
        setHistory((h) => [...h, res.player_word!]);
        setInput("");
        inputRef.current?.focus();
      } else if (!res.fatal) {
        setErrorMessage(res.message ?? "無効な入力です");
        setInput("");
      } else {
        if (res.player_word) {
          setHistory((h) => [...h, res.player_word!]);
        }
        setGameOver({ reason: res.reason });
      }
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "通信エラーが発生しました");
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) return <p className="text-zinc-500">準備中...</p>;

  if (gameOver) {
    return (
      <div className="space-y-4 text-center">
        <h2 className="text-xl font-bold text-zinc-900">ゲームオーバー</h2>
        <p className="text-zinc-600">使った単語数: {history.length}</p>
        <ol className="mx-auto max-w-xs space-y-1 text-left text-sm text-zinc-700">
          {history.map((w, i) => (
            <li key={i}>
              {i + 1}. {w.word}({w.reading})
            </li>
          ))}
        </ol>
        <div className="flex justify-center gap-4">
          <Link
            href="/play/solo"
            className="rounded bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700"
          >
            もう一度遊ぶ
          </Link>
          <Link
            href="/dashboard"
            className="rounded border border-zinc-300 px-4 py-2 text-zinc-700 hover:bg-zinc-100"
          >
            ダッシュボードへ
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 text-center">
      <div>
        <p className="text-sm text-zinc-500">直前の単語</p>
        <p className="text-3xl font-bold text-zinc-900">
          {lastWord ? lastWord.word : "最初の単語を入力してください"}
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit();
        }}
        className="flex justify-center gap-2"
      >
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="ひらがな・カタカナで入力"
          className="w-64 rounded border border-zinc-300 px-3 py-2 focus:border-zinc-500 focus:outline-none"
          autoFocus
        />
        <button
          type="submit"
          disabled={submitting}
          className="rounded bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700 disabled:opacity-50"
        >
          送信
        </button>
      </form>

      {errorMessage && <p className="text-sm text-red-600">{errorMessage}</p>}
    </div>
  );
}

export default function SoloPlayPage() {
  return (
    <RequireAuth>
      <div className="mx-auto flex w-full max-w-xl flex-1 flex-col justify-center px-4 py-8">
        <h1 className="mb-8 text-center text-2xl font-bold text-zinc-900">
          ソロプレイ
        </h1>
        <SoloPlay />
      </div>
    </RequireAuth>
  );
}
