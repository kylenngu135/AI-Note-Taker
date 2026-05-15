import type { Metadata } from "next";
import { Instrument_Serif, DM_Mono } from "next/font/google";
import { AuthProvider } from "@/context/AuthContext";
import "./globals.css";

const instrumentSerif = Instrument_Serif({
  weight: ["400"],
  style: ["normal", "italic"],
  subsets: ["latin"],
  variable: "--font-instrument-serif",
});

const dmMono = DM_Mono({
  weight: ["300", "400", "500"],
  subsets: ["latin"],
  variable: "--font-dm-mono",
});

export const metadata: Metadata = {
  title: "AI Notes Generator",
  description: "Upload documents or audio/video to generate AI study notes",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${instrumentSerif.variable} ${dmMono.variable} h-full overflow-hidden`}
    >
      <body className="h-full overflow-hidden bg-app-bg text-primary font-mono text-sm leading-relaxed">
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
