import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  experimental: {
    authInterrupts: true,
    serverActions: {
      bodySizeLimit: "8mb",
    },
  },
};

export default nextConfig;
