"use client";

import { Bot, ChevronRight, MessageCircle, ScrollText, User } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { ApiError, listSessions } from "@/lib/api";
import type { Session } from "@/lib/types";

const MODE_META = {
  solo: { label: "ソロプレイ", icon: User },
  vs_ai: { label: "AI対戦", icon: Bot },
  free_talk: { label: "フリートーク", icon: MessageCircle },
} as const;

const RESULT_LABELS: Record<string, string> = {
  win: "勝ち",
  lose: "負け",
  ended_by_n: "「ん」で終了",
  ended_by_duplicate: "重複により終了",
  ended_by_user: "自分で終了",
};

function statusBadgeClass(session: Session) {
  if (session.status === "in_progress") return "border-blue-200 bg-blue-50 text-blue-700";
  if (session.result === "win") return "border-green-200 bg-green-50 text-green-700";
  if (session.result === "lose") return "border-red-200 bg-red-50 text-red-700";
  return "border-border bg-muted text-muted-foreground";
}

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

  if (error) return <p className="text-sm text-destructive">{error}</p>;
  if (sessions === null) return <p className="text-sm text-muted-foreground">読み込み中...</p>;
  if (sessions.length === 0) {
    return <p className="text-sm text-muted-foreground">まだプレイ履歴がありません。</p>;
  }

  return (
    <div className="space-y-2">
      {sessions.map((s) => {
        const meta = MODE_META[s.mode];
        const Icon = meta?.icon ?? User;
        return (
          <Link key={s.id} href={`/history/${s.id}`}>
            <Card className="flex-row items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/50">
              <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-muted">
                <Icon className="size-4" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{meta?.label ?? s.mode}</p>
                <p className="text-xs text-muted-foreground">
                  {new Date(s.started_at).toLocaleString("ja-JP")}
                </p>
              </div>
              <Badge variant="outline" className={cn(statusBadgeClass(s))}>
                {s.status === "in_progress"
                  ? "プレイ中"
                  : s.result
                    ? (RESULT_LABELS[s.result] ?? s.result)
                    : "終了"}
              </Badge>
              <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
            </Card>
          </Link>
        );
      })}
    </div>
  );
}

export default function HistoryPage() {
  return (
    <RequireAuth>
      <div className="min-h-[calc(100vh-3.5rem)] bg-gradient-to-br from-slate-50 to-slate-100 px-4 py-12">
        <div className="mx-auto w-full max-w-xl space-y-6">
          <div className="flex items-center gap-2">
            <ScrollText className="size-5" />
            <h1 className="text-xl font-semibold">履歴</h1>
          </div>
          <HistoryList />
        </div>
      </div>
    </RequireAuth>
  );
}
