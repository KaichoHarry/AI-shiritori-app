"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";

import { ApiError, requestPasswordReset } from "@/lib/api";

export default function PasswordResetPage() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await requestPasswordReset(email);
      setDone(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "申請に失敗しました");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex flex-1 items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-6">
        <h1 className="text-center text-2xl font-bold text-zinc-900">
          パスワード再設定
        </h1>

        {done ? (
          <div className="space-y-4 text-center text-sm text-zinc-700">
            <p>
              申請を受け付けました。管理者の承認をお待ちください。承認され次第、登録済みのメールアドレスに新しいパスワードが届きます。
            </p>
            <Link href="/login" className="text-zinc-900 underline">
              ログイン画面に戻る
            </Link>
          </div>
        ) : (
          <>
            <p className="text-sm text-zinc-600">
              登録済みのメールアドレスを入力してください。管理者の承認後、新しいパスワードがメールで届きます。
            </p>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-zinc-700">
                  メールアドレス
                </label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full rounded border border-zinc-300 px-3 py-2 focus:border-zinc-500 focus:outline-none"
                />
              </div>

              {error && <p className="text-sm text-red-600">{error}</p>}

              <button
                type="submit"
                disabled={submitting}
                className="w-full rounded bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700 disabled:opacity-50"
              >
                {submitting ? "送信中..." : "申請する"}
              </button>
            </form>
            <p className="text-center text-sm text-zinc-600">
              <Link href="/login" className="hover:underline">
                ログイン画面に戻る
              </Link>
            </p>
          </>
        )}
      </div>
    </div>
  );
}
