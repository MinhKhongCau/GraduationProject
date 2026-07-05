import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  env: {
    REACT_APP_API_URL: process.env.REACT_APP_API_URL,
  },
};

export default nextConfig;
