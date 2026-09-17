import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  images: {
    // 静的出力(Cloudflare Pages)ではNext.jsの画像最適化サーバーが使えないため無効化する
    unoptimized: true,
  },
};

export default nextConfig;
