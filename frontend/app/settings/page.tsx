"use client";

import { Check, MessageSquareHeart, Save, Settings as SettingsIcon, Swords } from "lucide-react";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { ApiError, getSettings, updateSettings } from "@/lib/api";
import type { Difficulty, Tone } from "@/lib/types";

const DIFFICULTY_LABELS: Record<Difficulty, string> = {
  easy: "易しい",
  normal: "普通",
  hard: "難しい",
};

const TONE_LABELS: Record<Tone, string> = {
  friendly: "フレンドリー",
  polite: "丁寧",
  comedy: "コント風",
};

function OptionButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-sm transition-colors",
        active
          ? "border-primary bg-primary text-primary-foreground"
          : "border-border text-muted-foreground hover:bg-muted",
      )}
    >
      {active && <Check className="size-3.5" />}
      {children}
    </button>
  );
}

function SettingsForm() {
  const { authFetch } = useAuth();
  const [difficulty, setDifficulty] = useState<Difficulty>("normal");
  const [tone, setTone] = useState<Tone>("friendly");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    authFetch((token) => getSettings(token))
      .then((s) => {
        setDifficulty(s.mode2_difficulty);
        setTone(s.mode3_tone);
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "設定の取得に失敗しました");
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function handleSave() {
    setSaving(true);
    setSaved(false);
    setError(null);
    try {
      await authFetch((token) => updateSettings(token, difficulty, tone));
      setSaved(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "設定の更新に失敗しました");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-muted-foreground">読み込み中...</p>;
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="flex-row items-center gap-2 space-y-0">
          <Swords className="size-4 text-muted-foreground" />
          <CardTitle className="text-sm font-medium">
            AI対戦のデフォルト難易度
          </CardTitle>
        </CardHeader>
        <CardContent className="flex gap-2">
          {(Object.keys(DIFFICULTY_LABELS) as Difficulty[]).map((d) => (
            <OptionButton key={d} active={difficulty === d} onClick={() => setDifficulty(d)}>
              {DIFFICULTY_LABELS[d]}
            </OptionButton>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex-row items-center gap-2 space-y-0">
          <MessageSquareHeart className="size-4 text-muted-foreground" />
          <CardTitle className="text-sm font-medium">
            フリートークの会話トーン
          </CardTitle>
        </CardHeader>
        <CardContent className="flex gap-2">
          {(Object.keys(TONE_LABELS) as Tone[]).map((t) => (
            <OptionButton key={t} active={tone === t} onClick={() => setTone(t)}>
              {TONE_LABELS[t]}
            </OptionButton>
          ))}
        </CardContent>
      </Card>

      <div className="flex items-center gap-3">
        <Button onClick={handleSave} disabled={saving}>
          <Save />
          {saving ? "保存中..." : "保存する"}
        </Button>
        {error && <p className="text-sm text-destructive">{error}</p>}
        {saved && <p className="text-sm text-green-600">保存しました</p>}
      </div>
    </div>
  );
}

export default function SettingsPage() {
  return (
    <RequireAuth>
      <div className="min-h-[calc(100vh-3.5rem)] bg-gradient-to-br from-slate-50 to-slate-100 px-4 py-12">
        <div className="mx-auto w-full max-w-xl space-y-6">
          <div className="flex items-center gap-2">
            <SettingsIcon className="size-5" />
            <h1 className="text-xl font-semibold">設定</h1>
          </div>
          <SettingsForm />
        </div>
      </div>
    </RequireAuth>
  );
}
