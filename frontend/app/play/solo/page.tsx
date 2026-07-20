"use client";

import { Send, Trophy, User } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { ApiError, createSession, submitWord } from "@/lib/api";
import type { WordView } from "@/lib/types";

function SoloPlay() {
  const { authFetch } = useAuth();
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [lastWord, setLastWord] = useState<WordView | null>(null);
  const [history, setHistory] = useState<WordView[]>([]);
  const [input, setInput] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [gameOver, setGameOver] = useState(false);
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
        setGameOver(true);
      }
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "通信エラーが発生しました");
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <p className="text-center text-sm text-muted-foreground">準備中...</p>
    );
  }

  if (gameOver) {
    return (
      <div className="mx-auto w-full max-w-sm rounded-2xl bg-white p-10 text-center shadow-sm ring-1 ring-black/5">
        <Trophy className="mx-auto mb-4 size-10 text-muted-foreground" />
        <h2 className="text-xl font-bold">ゲームオーバー</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          使った単語数: {history.length}
        </p>
        <ol className="mt-6 space-y-2 text-left text-sm">
          {history.map((w, i) => (
            <li key={i} className="text-muted-foreground">
              {i + 1}. <span className="text-foreground">{w.word}</span>
              <span className="ml-1">({w.reading})</span>
            </li>
          ))}
        </ol>
        <div className="mt-8 flex justify-center gap-3">
          <Link href="/play/solo" className={cn(buttonVariants(), "h-10")}>
            もう一度遊ぶ
          </Link>
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

  return (
    <div className="mx-auto w-full max-w-md space-y-10">
      <div className="rounded-2xl bg-white p-10 text-center shadow-sm ring-1 ring-black/5">
        <p className="flex items-center justify-center gap-1.5 text-xs text-muted-foreground">
          <User className="size-3.5" />
          直前の単語
        </p>
        <p className="mt-4 text-3xl font-bold">
          {lastWord ? lastWord.word : "最初の単語をどうぞ"}
        </p>
      </div>

      <div className="space-y-3">
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

export default function SoloPlayPage() {
  return (
    <RequireAuth>
      <div className="flex min-h-[calc(100vh-3.5rem)] flex-col justify-center bg-gradient-to-br from-slate-50 to-slate-100 px-4 py-16">
        <h1 className="mb-10 text-center text-xl font-semibold">ソロプレイ</h1>
        <SoloPlay />
      </div>
    </RequireAuth>
  );
}
