"use client";

import Link from "next/link";
import { useRef, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { ApiError, createSession, submitWord } from "@/lib/api";
import type { Difficulty, WordView } from "@/lib/types";

type Turn = { speaker: "user" | "ai"; word: WordView };

const DIFFICULTY_LABELS: Record<Difficulty, string> = {
  easy: "易しい",
  normal: "普通",
  hard: "難しい",
};

function VsAiPlay() {
  const { authFetch } = useAuth();
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [difficulty, setDifficulty] = useState<Difficulty | null>(null);
  const [turns, setTurns] = useState<Turn[]>([]);
  const [input, setInput] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [gameOver, setGameOver] = useState<{
    result?: string;
    reason?: string;
  } | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  async function start(selected: Difficulty) {
    setErrorMessage(null);
    try {
      const s = await authFetch((token) => createSession(token, "vs_ai", selected));
      setSessionId(s.id);
      setDifficulty(selected);
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "開始に失敗しました");
    }
  }

  async function handleSubmit() {
    if (!sessionId || !input.trim() || submitting) return;
    setSubmitting(true);
    setErrorMessage(null);
    try {
      const res = await authFetch((token) => submitWord(token, sessionId, input));

      if (res.player_word) {
        setTurns((t) => [...t, { speaker: "user", word: res.player_word! }]);
      }
      if (res.ai_word) {
        setTurns((t) => [...t, { speaker: "ai", word: res.ai_word! }]);
      }

      if (!res.accepted && !res.fatal) {
        setErrorMessage(res.message ?? "無効な入力です");
      } else {
        setInput("");
        inputRef.current?.focus();
      }

      if (res.status === "finished") {
        setGameOver({ result: res.result, reason: res.reason });
      }
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "通信エラーが発生しました");
    } finally {
      setSubmitting(false);
    }
  }

  if (!sessionId || !difficulty) {
    return (
      <div className="space-y-4 text-center">
        <p className="text-zinc-600">難易度を選んでください</p>
        <div className="flex justify-center gap-2">
          {(Object.keys(DIFFICULTY_LABELS) as Difficulty[]).map((d) => (
            <button
              key={d}
              type="button"
              onClick={() => start(d)}
              className="rounded border border-zinc-300 px-4 py-2 hover:bg-zinc-100"
            >
              {DIFFICULTY_LABELS[d]}
            </button>
          ))}
        </div>
        {errorMessage && <p className="text-sm text-red-600">{errorMessage}</p>}
      </div>
    );
  }

  if (gameOver) {
    const won = gameOver.result === "win";
    return (
      <div className="space-y-4 text-center">
        <h2 className="text-xl font-bold text-zinc-900">
          {won ? "あなたの勝ちです！" : "あなたの負けです"}
        </h2>
        <ol className="mx-auto max-w-xs space-y-1 text-left text-sm text-zinc-700">
          {turns.map((t, i) => (
            <li key={i}>
              {i + 1}. [{t.speaker === "ai" ? "AI" : "自分"}] {t.word.word}
            </li>
          ))}
        </ol>
        <div className="flex justify-center gap-4">
          <Link
            href="/play/vs-ai"
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

  const lastTurn = turns[turns.length - 1];

  return (
    <div className="space-y-6 text-center">
      <div>
        <p className="text-sm text-zinc-500">直前のやり取り</p>
        <p className="text-3xl font-bold text-zinc-900">
          {lastTurn ? lastTurn.word.word : "最初の単語を入力してください"}
        </p>
        {lastTurn && (
          <p className="text-xs text-zinc-500">
            ({lastTurn.speaker === "ai" ? "AI" : "自分"})
          </p>
        )}
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

export default function VsAiPlayPage() {
  return (
    <RequireAuth>
      <div className="mx-auto flex w-full max-w-xl flex-1 flex-col justify-center px-4 py-8">
        <h1 className="mb-8 text-center text-2xl font-bold text-zinc-900">
          AI対戦
        </h1>
        <VsAiPlay />
      </div>
    </RequireAuth>
  );
}
