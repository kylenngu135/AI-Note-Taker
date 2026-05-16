import type { Metadata } from 'next';
import { DM_Sans, DM_Serif_Display, JetBrains_Mono } from 'next/font/google';
import '@/styles/globals.css';
import { ToastProvider } from '@/components/ui/Toast';

const dmSans = DM_Sans({
  subsets: ['latin'],
  variable: '--font-dm-sans',
  weight: ['400', '500', '600'],
  display: 'swap',
});

const dmSerifDisplay = DM_Serif_Display({
  subsets: ['latin'],
  variable: '--font-dm-serif-display',
  weight: ['400'],
  display: 'swap',
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-jetbrains-mono',
  weight: ['400'],
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Note Taker',
  description: 'AI-powered study notes from your documents',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      className={`${dmSans.variable} ${dmSerifDisplay.variable} ${jetbrainsMono.variable}`}
    >
      <body>
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
