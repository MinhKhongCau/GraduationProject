import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { cookies } from "next/headers";
import { AppProviders } from "@/context";
import { ServiceWorkerRegister } from "@/components/pwa/ServiceWorkerRegister";
import { LOCALE_COOKIE } from "@/constants/api";
import type { Locale } from "@/context/LocaleContext";
import "./index.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "MindCare",
  description: "Nền tảng kết nối chuyên gia tâm lý và khách hàng hàng đầu. Tại MindCare, chúng tôi cam kết mang đến trải nghiệm tư vấn tâm lý trực tuyến an toàn, tiện lợi và hiệu quả. Với đội ngũ chuyên gia tâm lý giàu kinh nghiệm và đa dạng chuyên môn, chúng tôi giúp bạn vượt qua những khó khăn trong cuộc sống, cải thiện sức khỏe tinh thần và đạt được sự cân bằng trong cuộc sống hàng ngày.",
  icons: {
    icon: [
      { url: "/icons/icon-192.png", sizes: "192x192", type: "image/png" },
      { url: "/icons/icon-512.png", sizes: "512x512", type: "image/png" },
    ],
    apple: [{ url: "/icons/apple-touch-icon.png", sizes: "180x180" }],
  },
  appleWebApp: {
    capable: true,
    title: "MindCare",
    statusBarStyle: "default",
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
  themeColor: "#0369a1",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const cookieLocale = cookieStore.get(LOCALE_COOKIE)?.value;
  const initialLocale: Locale = cookieLocale === "en" ? "en" : "vi";

  return (
    <html
      lang={initialLocale}
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <AppProviders initialLocale={initialLocale}>{children}</AppProviders>
        <ServiceWorkerRegister />
      </body>
    </html>
  );
}
