import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { cookies } from "next/headers";
import { AppProviders } from "@/context";
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
    // Kỹ thuật dùng Emoji làm Favicon thông qua SVG
    icon: `data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🧠</text></svg>`,
  },
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
      </body>
    </html>
  );
}
