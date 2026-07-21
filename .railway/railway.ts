import { defineRailway, github, postgres, project, service, volume } from "railway/iac";

export default defineRailway(() => {
  const Postgres = postgres("Postgres", { region: "sfo" });
  const postgresVolume = volume("postgres-volume", { alerts: { usage: { "100": {}, "80": {}, "95": {} } }, allowOnlineResize: true, region: "sfo", sizeMB: 500 });

  // 秘密情報(JWT_ACCESS_SECRET, JWT_REFRESH_SECRET, GEMINI_API_KEY, SMTP_*, IMAP_*等)は
  // ここには書かず、`railway variable set`で個別に設定する(READMEのパスワード再設定の
  // 設計判断どおり、秘密情報はコード管理外にする方針)。
  const backend = service("backend", {
    source: github("KaichoHarry/AI-shiritori-app", { branch: "main", rootDirectory: "backend" }),
    env: {
      DATABASE_URL: Postgres.env.DATABASE_URL,
      GEMINI_MODEL: "gemini-3.5-flash",
      IMAP_POLL_INTERVAL_SECONDS: "30",
    },
  });

  const frontend = service("frontend", {
    source: github("KaichoHarry/AI-shiritori-app", { branch: "main", rootDirectory: "frontend" }),
  });

  return project("ai-shiritori-app", {
    resources: [Postgres, postgresVolume, backend, frontend],
  });
});
