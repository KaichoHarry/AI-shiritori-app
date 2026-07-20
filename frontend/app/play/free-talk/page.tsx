"use client";

import { Bot, Send, Square, User } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
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
  const [finished, setFinished] = useState(false);
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
        setFinished(true);
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
      setFinished(true);
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "終了に失敗しました");
    }
  }

  if (loading) {
    return (
      <p className="text-center text-sm text-muted-foreground">準備中...</p>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6">
      <div className="flex flex-1 flex-col gap-4 overflow-y-auto rounded-2xl bg-white p-6 shadow-sm ring-1 ring-black/5">
        {messages.length === 0 && (
          <p className="pt-12 text-center text-sm text-muted-foreground">
            最初のメッセージを送ってみましょう
            <br />
            (ひらがな・カタカナ・英字で終わる文にしてください)
          </p>
        )}
        {messages.map((m, i) => (
          <div
            key={i}
            className={cn(
              "flex items-end gap-2",
              m.speaker === "user" ? "justify-end" : "justify-start",
            )}
          >
            {m.speaker === "ai" && (
              <Bot className="mb-1 size-5 shrink-0 text-blue-600" />
            )}
            <div
              className={cn(
                "max-w-[75%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed",
                m.speaker === "user"
                  ? "rounded-br-md bg-primary text-primary-foreground"
                  : "rounded-bl-md bg-blue-50 text-foreground",
              )}
            >
              {m.content}
            </div>
            {m.speaker === "user" && (
              <User className="mb-1 size-5 shrink-0 text-muted-foreground" />
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {finished ? (
        <div className="rounded-2xl bg-white p-8 text-center shadow-sm ring-1 ring-black/5">
          <p className="text-sm font-medium">会話が終了しました</p>
          <div className="mt-5 flex justify-center gap-3">
            <Link href="/play/free-talk" className={cn(buttonVariants(), "h-10")}>
              もう一度話す
            </Link>
            <Link
              href="/dashboard"
              className={cn(buttonVariants({ variant: "outline" }), "h-10")}
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
          className="flex gap-2"
        >
          <Input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="メッセージを入力"
            className="h-11"
            autoFocus
          />
          <Button type="submit" disabled={submitting} size="icon-lg" aria-label="送信">
            <Send />
          </Button>
          <Button
            type="button"
            variant="outline"
            size="icon-lg"
            aria-label="終了する"
            onClick={handleEnd}
          >
            <Square />
          </Button>
        </form>
      )}

      {errorMessage && (
        <p className="text-center text-sm text-destructive">{errorMessage}</p>
      )}
    </div>
  );
}

export default function FreeTalkPlayPage() {
  return (
    <RequireAuth>
      <div className="flex min-h-[calc(100vh-3.5rem)] flex-col bg-gradient-to-br from-slate-50 to-slate-100 px-4 py-10">
        <h1 className="mb-6 text-center text-xl font-semibold">フリートーク</h1>
        <FreeTalkPlay />
      </div>
    </RequireAuth>
  );
}
