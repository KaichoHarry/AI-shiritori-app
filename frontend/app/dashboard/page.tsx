"use client";

import { Bot, ListChecks, MessageCircle, User } from "lucide-react";
import Link from "next/link";

import { useAuth } from "@/components/AuthProvider";
import { RequireAuth } from "@/components/RequireAuth";

const CARDS = [
  {
    href: "/play/solo",
    title: "ソロプレイ",
    description: "一人でしりとりを続けよう",
    icon: User,
    iconBg: "bg-slate-900",
  },
  {
    href: "/play/vs-ai",
    title: "AI対戦",
    description: "AIと交互に単語を出し合う",
    icon: Bot,
    iconBg: "bg-blue-600",
  },
  {
    href: "/play/free-talk",
    title: "フリートーク",
    description: "AIと文章しりとりで会話する",
    icon: MessageCircle,
    iconBg: "bg-rose-500",
  },
];

const COMMON_RULES = [
  "直前の言葉の最後の音(モーラ)から始まる言葉をつなげます。",
  "「ん」で終わる言葉を出すと、その時点で負けです。",
  "同じセッション内で一度使った言葉は、もう一度使えません。",
  "ひらがな・カタカナのみで入力します(漢字・英数字は使えません。フリートークの会話文は除く)。",
  "長音「ー」は直前の音の母音として扱われます(例:「コーヒー」の次は「ひ」から始まる言葉)。",
  "拗音「ゃ・ゅ・ょ」や、外来語表記で使う小さい母音「ぁ・ぃ・ぅ・ぇ・ぉ」(例:「ファ・ティ・チェ」)は、直前の文字と合わせて1つの音として扱われます(例:「カフェ」の次は「ふぇ」から始まる言葉)。",
  "促音「っ」はこれらとは異なり、それ自体で独立した1つの音として扱われます(例:「らっぱ」の次は「ぱ」から始まる言葉)。",
];

const MODE_GUIDES = [
  {
    title: "ソロプレイ",
    icon: User,
    iconBg: "bg-slate-900",
    points: [
      "一人でどこまで言葉をつなげられるかに挑戦するモードです。共通ルールに反する言葉(辞書にない言葉・「ん」で終わる言葉・既に使った言葉)を入力すると終了です。",
      "入力欄の上に、次に入力すべき音のヒントがひらがな・カタカナ両方で表示されます。",
    ],
  },
  {
    title: "AI対戦",
    icon: Bot,
    iconBg: "bg-blue-600",
    points: [
      "AIと交互に言葉を出し合い、先に共通ルールに反した方が負けです。",
      "難易度は「易しい」「普通」「難しい」の3段階。難易度が上がるほどAIは手強くなります(詳しい仕組みはプレイしてのお楽しみです)。",
    ],
  },
  {
    title: "フリートーク(モード3)",
    icon: MessageCircle,
    iconBg: "bg-rose-500",
    points: [
      "単語単位ではなく、AIと「文章」でしりとりをしながら会話を続けるモードです。他の2モードとはルールが異なるので注意してください。",
      "直前の発言(自分の発言でもAIの発言でも構いません)の文末の音から、次の発言を書き始めます。単語のしりとりのような厳密さは不要で、自然な文章のつながりを重視します。",
      "文末が「。」「!」のような句読点・記号・絵文字で終わっていても構いません。それらは無視され、その手前にある最後の有効な文字(ひらがな・カタカナ・半角/全角英字)まで遡って次の音が決まります。漢字も同様に読み飛ばされる扱いで、特別扱いはされません(送り仮名のない漢字だけで文が終わり、それより前に有効な文字が全く見つからない場合のみ、次の音を判定できません。通常の文章ではまず起こりません)。",
      "発言が「ん」の音で終わると、その時点で会話は終了します。",
      "AIの返事を書いている間は、吹き出しに「AIが返事を考えています」という表示が出ます。",
      "入力欄の上に、次に何の音から始めればよいかのヒントが表示されます。",
      "AIの口調(フレンドリー・丁寧・コント風)は「設定」画面から変更できます。途中でやめたいときは、送信欄の隣の■ボタンでいつでも会話を終了できます。",
    ],
  },
];

function DashboardContent() {
  const { user } = useAuth();

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <main className="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="mb-10 rounded-2xl bg-white p-8 shadow-sm ring-1 ring-black/5">
          <h1 className="text-2xl font-bold tracking-tight">
            おかえりなさい、{user?.display_name}さん
          </h1>
          <p className="mt-2 text-muted-foreground">
            モードを選んでプレイを始めましょう
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {CARDS.map(({ href, title, description, icon: Icon, iconBg }) => (
            <Link key={href} href={href}>
              <div className="flex h-full flex-col gap-4 rounded-2xl bg-white p-6 shadow-sm ring-1 ring-black/5 transition-shadow hover:shadow-md">
                <div
                  className={`flex size-12 items-center justify-center rounded-full text-white ${iconBg}`}
                >
                  <Icon className="size-6" />
                </div>
                <div>
                  <h2 className="font-semibold">{title}</h2>
                  <p className="mt-1 text-sm text-muted-foreground">{description}</p>
                </div>
              </div>
            </Link>
          ))}
        </div>

        <div className="mt-10 space-y-6">
          <div className="rounded-2xl bg-white p-8 shadow-sm ring-1 ring-black/5">
            <div className="mb-4 flex items-center gap-2">
              <ListChecks className="size-5 text-slate-700" />
              <h2 className="text-lg font-semibold">共通ルール</h2>
            </div>
            <ul className="space-y-2 text-sm text-muted-foreground">
              {COMMON_RULES.map((rule, i) => (
                <li key={i} className="flex gap-2">
                  <span className="text-slate-400">・</span>
                  <span>{rule}</span>
                </li>
              ))}
            </ul>
          </div>

          {MODE_GUIDES.map(({ title, icon: Icon, iconBg, points }) => (
            <div
              key={title}
              className="rounded-2xl bg-white p-8 shadow-sm ring-1 ring-black/5"
            >
              <div className="mb-4 flex items-center gap-3">
                <div
                  className={`flex size-9 items-center justify-center rounded-full text-white ${iconBg}`}
                >
                  <Icon className="size-4.5" />
                </div>
                <h2 className="text-lg font-semibold">{title}</h2>
              </div>
              <ul className="space-y-2 text-sm text-muted-foreground">
                {points.map((point, i) => (
                  <li key={i} className="flex gap-2">
                    <span className="text-slate-400">・</span>
                    <span>{point}</span>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}

export default function DashboardPage() {
  return (
    <RequireAuth>
      <DashboardContent />
    </RequireAuth>
  );
}
