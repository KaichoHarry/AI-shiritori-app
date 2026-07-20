"use client";

import { useEffect, useState } from "react";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";
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
    return <p className="text-zinc-500">読み込み中...</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="mb-2 text-lg font-semibold text-zinc-900">
          AI対戦のデフォルト難易度
        </h2>
        <div className="flex gap-2">
          {(Object.keys(DIFFICULTY_LABELS) as Difficulty[]).map((d) => (
            <button
              key={d}
              type="button"
              onClick={() => setDifficulty(d)}
              className={`rounded border px-3 py-1.5 text-sm ${
                difficulty === d
                  ? "border-zinc-900 bg-zinc-900 text-white"
                  : "border-zinc-300 text-zinc-700 hover:bg-zinc-100"
              }`}
            >
              {DIFFICULTY_LABELS[d]}
            </button>
          ))}
        </div>
      </div>

      <div>
        <h2 className="mb-2 text-lg font-semibold text-zinc-900">
          フリートークの会話トーン
        </h2>
        <div className="flex gap-2">
          {(Object.keys(TONE_LABELS) as Tone[]).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTone(t)}
              className={`rounded border px-3 py-1.5 text-sm ${
                tone === t
                  ? "border-zinc-900 bg-zinc-900 text-white"
                  : "border-zinc-300 text-zinc-700 hover:bg-zinc-100"
              }`}
            >
              {TONE_LABELS[t]}
            </button>
          ))}
        </div>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}
      {saved && <p className="text-sm text-green-600">保存しました</p>}

      <button
        type="button"
        onClick={handleSave}
        disabled={saving}
        className="rounded bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700 disabled:opacity-50"
      >
        {saving ? "保存中..." : "保存する"}
      </button>
    </div>
  );
}

export default function SettingsPage() {
  return (
    <RequireAuth>
      <div className="mx-auto w-full max-w-2xl px-4 py-8">
        <h1 className="mb-6 text-2xl font-bold text-zinc-900">設定</h1>
        <SettingsForm />
      </div>
    </RequireAuth>
  );
}
