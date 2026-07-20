"use client";

import { KeyRound, Mail, Send } from "lucide-react";
import Link from "next/link";
import { useState, type FormEvent } from "react";

import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
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
    <div className="flex min-h-screen w-full items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 p-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
            <KeyRound className="size-8" />
          </div>
          <h1 className="text-4xl font-bold tracking-tight">パスワード再設定</h1>
        </div>

        <div className="rounded-2xl bg-white/95 p-8 shadow-2xl ring-1 ring-black/5 backdrop-blur-sm">
          {done ? (
            <div className="space-y-6 text-center text-sm text-muted-foreground">
              <p>
                申請を受け付けました。管理者の承認をお待ちください。承認され次第、登録済みのメールアドレスに新しいパスワードが届きます。
              </p>
              <Link
                href="/login"
                className={cn(buttonVariants({ variant: "outline" }), "h-11 w-full")}
              >
                ログイン画面に戻る
              </Link>
            </div>
          ) : (
            <div className="space-y-6">
              <p className="text-sm text-muted-foreground">
                登録済みのメールアドレスを入力してください。管理者の承認後、新しいパスワードがメールで届きます。
              </p>
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="email">メールアドレス</Label>
                  <div className="relative">
                    <Mail className="pointer-events-none absolute top-1/2 left-3 size-5 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      id="email"
                      type="email"
                      required
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="h-11 pl-10"
                      placeholder="example@email.com"
                    />
                  </div>
                </div>

                {error && <p className="text-sm text-destructive">{error}</p>}

                <div className="space-y-3">
                  <Button type="submit" disabled={submitting} className="h-11 w-full">
                    <Send />
                    {submitting ? "送信中..." : "申請する"}
                  </Button>
                  <Link
                    href="/login"
                    className={cn(buttonVariants({ variant: "outline" }), "h-11 w-full")}
                  >
                    ログイン画面に戻る
                  </Link>
                </div>
              </form>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
