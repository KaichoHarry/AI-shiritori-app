"use client";

import { LayoutGrid, LogOut, ScrollText, Settings } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { useAuth } from "@/components/AuthProvider";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const NAV_ITEMS = [
  { href: "/dashboard", label: "ダッシュボード", icon: LayoutGrid },
  { href: "/settings", label: "設定", icon: Settings },
  { href: "/history", label: "履歴", icon: ScrollText },
];

export function Header() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  if (!user) return null;

  return (
    <header className="sticky top-0 z-10 border-b bg-background/95 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-3xl items-center justify-between px-4">
        <Link href="/dashboard" className="text-base font-semibold tracking-tight">
          しりとりアプリ
        </Link>

        <nav className="flex items-center gap-1">
          {NAV_ITEMS.map(({ href, label, icon: Icon }) => {
            const active = pathname === href;
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  buttonVariants({
                    variant: active ? "secondary" : "ghost",
                    size: "sm",
                  }),
                  !active && "text-muted-foreground",
                )}
              >
                <Icon />
                <span className="hidden sm:inline">{label}</span>
              </Link>
            );
          })}
          <Button
            variant="ghost"
            size="icon"
            aria-label="ログアウト"
            onClick={() => {
              logout();
              router.replace("/login");
            }}
          >
            <LogOut />
          </Button>
        </nav>
      </div>
    </header>
  );
}
