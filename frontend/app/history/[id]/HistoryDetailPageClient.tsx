"use client";

import { ArrowLeft, Bot, ScrollText, User } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ApiError, getSessionDetail } from "@/lib/api";
import type { MessageRecord, Session, WordRecord } from "@/lib/types";

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

  if (error) return <p className="text-sm text-destructive">{error}</p>;
  if (!session) return <p className="text-sm text-muted-foreground">読み込み中...</p>;

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">{MODE_LABELS[session.mode] ?? session.mode}</Badge>
          <Badge variant="outline">
            {session.status === "in_progress" ? "プレイ中" : "終了"}
          </Badge>
          {session.result && (
            <Badge variant="outline">
              {RESULT_LABELS[session.result] ?? session.result}
            </Badge>
          )}
          <span className="text-sm text-muted-foreground">
            ターン数: {session.turn_count}
          </span>
        </CardContent>
      </Card>

      {words && (
        <ol className="space-y-2">
          {words.map((w) => (
            <li key={w.seq_no}>
              <Card
                className={cn(
                  "flex-row items-center gap-3 px-3 py-2",
                  w.speaker === "ai" && "bg-blue-50",
                )}
              >
                {w.speaker === "ai" ? (
                  <Bot className="size-4 shrink-0 text-blue-600" />
                ) : (
                  <User className="size-4 shrink-0 text-muted-foreground" />
                )}
                <span className="text-sm">
                  {w.word}
                  <span className="ml-1 text-muted-foreground">({w.reading})</span>
                </span>
              </Card>
            </li>
          ))}
        </ol>
      )}

      {messages && (
        <ol className="space-y-2">
          {messages.map((m) => (
            <li key={m.seq_no}>
              <Card
                className={cn(
                  "flex-row items-start gap-3 px-3 py-2.5",
                  m.speaker === "ai" && "bg-blue-50",
                )}
              >
                {m.speaker === "ai" ? (
                  <Bot className="mt-0.5 size-4 shrink-0 text-blue-600" />
                ) : (
                  <User className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                )}
                <span className="text-sm">{m.content}</span>
              </Card>
            </li>
          ))}
        </ol>
      )}

      <Link
        href="/history"
        className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
      >
        <ArrowLeft />
        履歴一覧に戻る
      </Link>
    </div>
  );
}

export default function HistoryDetailPageClient() {
  return (
    <RequireAuth>
      <div className="min-h-[calc(100vh-3.5rem)] bg-gradient-to-br from-slate-50 to-slate-100 px-4 py-12">
        <div className="mx-auto w-full max-w-xl space-y-6">
          <div className="flex items-center gap-2">
            <ScrollText className="size-5" />
            <h1 className="text-xl font-semibold">履歴詳細</h1>
          </div>
          <HistoryDetail />
        </div>
      </div>
    </RequireAuth>
  );
}
