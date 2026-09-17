import HistoryDetailPageClient from "./HistoryDetailPageClient";

// 静的エクスポート(output: "export")では動的ルートに generateStaticParams が必須。
// また、Next.js 16.2.10時点では空配列 [] を返すと(実際には定義されているにもかかわらず)
// 「generateStaticParams()がありません」という誤ったビルドエラーになる既知の挙動があるため、
// ダミーの1件だけを返す。実際のIDごとのページは事前生成せず、Cloudflare側のSPAフォールバック
// (not_found_handling: "single-page-application")によりどのIDでもこのアプリのシェルが
// 返されるようにし、実際の描画はクライアント側で useParams() を使って行う。
export function generateStaticParams() {
  return [{ id: "placeholder" }];
}

export default function HistoryDetailPage() {
  return <HistoryDetailPageClient />;
}
