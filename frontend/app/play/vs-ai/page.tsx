"use client";

import { Bot, PartyPopper, Send, User } from "lucide-react";
import Link from "next/link";
import { useRef, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { ApiError, createSession, submitWord } from "@/lib/api";
import type { Difficulty, NextHint, WordView } from "@/lib/types";

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
  const [nextHint, setNextHint] = useState<NextHint | null>(null);
  const [input, setInput] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [gameOver, setGameOver] = useState<{ result?: string } | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  function resetGame() {
    setSessionId(null);
    setDifficulty(null);
    setTurns([]);
    setNextHint(null);
    setInput("");
    setErrorMessage(null);
    setGameOver(null);
  }

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

      setNextHint(res.next_hint ?? null);

      if (res.status === "finished") {
        setGameOver({ result: res.result });
      }
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "通信エラーが発生しました");
    } finally {
      setSubmitting(false);
    }
  }

  if (!sessionId || !difficulty) {
    return (
      <div className="mx-auto w-full max-w-sm rounded-2xl bg-white p-10 text-center shadow-sm ring-1 ring-black/5">
        <Bot className="mx-auto mb-4 size-10 text-muted-foreground" />
        <p className="text-sm text-muted-foreground">難易度を選んでください</p>
        <div className="mt-6 flex justify-center gap-3">
          {(Object.keys(DIFFICULTY_LABELS) as Difficulty[]).map((d) => (
            <Button key={d} variant="outline" className="h-10" onClick={() => start(d)}>
              {DIFFICULTY_LABELS[d]}
            </Button>
          ))}
        </div>
        {errorMessage && (
          <p className="mt-4 text-sm text-destructive">{errorMessage}</p>
        )}
      </div>
    );
  }

  if (gameOver) {
    const won = gameOver.result === "win";
    return (
      <div className="mx-auto w-full max-w-sm rounded-2xl bg-white p-10 text-center shadow-sm ring-1 ring-black/5">
        {won ? (
          <PartyPopper className="mx-auto mb-4 size-10 text-amber-500" />
        ) : (
          <Bot className="mx-auto mb-4 size-10 text-muted-foreground" />
        )}
        <h2 className="text-xl font-bold">
          {won ? "あなたの勝ちです！" : "あなたの負けです"}
        </h2>
        <ol className="mt-6 space-y-2 text-left text-sm">
          {turns.map((t, i) => (
            <li key={i} className="flex items-center gap-1.5 text-muted-foreground">
              {t.speaker === "ai" ? (
                <Bot className="size-3.5 shrink-0" />
              ) : (
                <User className="size-3.5 shrink-0" />
              )}
              <span className="text-foreground">{t.word.word}</span>
            </li>
          ))}
        </ol>
        <div className="mt-8 flex justify-center gap-3">
          <Button className="h-10" onClick={resetGame}>
            もう一度遊ぶ
          </Button>
          <Link
            href="/dashboard"
            className={cn(buttonVariants({ variant: "outline" }), "h-10")}
          >
            ダッシュボードへ
          </Link>
        </div>
      </div>
    );
  }

  const lastTurn = turns[turns.length - 1];

  return (
    <div className="mx-auto w-full max-w-md space-y-10">
      <div className="rounded-2xl bg-white p-10 text-center shadow-sm ring-1 ring-black/5">
        <p className="flex items-center justify-center gap-1.5 text-xs text-muted-foreground">
          {lastTurn?.speaker === "ai" ? (
            <Bot className="size-3.5" />
          ) : (
            <User className="size-3.5" />
          )}
          直前のやり取り
        </p>
        <p className="mt-4 text-3xl font-bold">
          {lastTurn ? lastTurn.word.word : "最初の単語をどうぞ"}
        </p>
      </div>

      <div className="space-y-3">
        {nextHint && (
          <p className="text-center text-sm text-muted-foreground">
            「{nextHint.hiragana}」または「{nextHint.katakana}」から始まる言葉を入力してください
          </p>
        )}

        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="flex gap-2"
        >
          <Input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="ひらがな・カタカナで入力"
            className="h-11"
            autoFocus
          />
          <Button type="submit" disabled={submitting} size="icon-lg" aria-label="送信">
            <Send />
          </Button>
        </form>

        {errorMessage && (
          <p className="text-center text-sm text-destructive">{errorMessage}</p>
        )}
      </div>
    </div>
  );
}

export default function VsAiPlayPage() {
  return (
    <RequireAuth>
      <div className="flex min-h-[calc(100vh-3.5rem)] flex-col justify-center bg-gradient-to-br from-slate-50 to-slate-100 px-4 py-16">
        <h1 className="mb-10 text-center text-xl font-semibold">AI対戦</h1>
        <VsAiPlay />
      </div>
    </RequireAuth>
  );
}
