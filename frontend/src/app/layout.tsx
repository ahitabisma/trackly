import type { Metadata } from "next";
import "./globals.css";
import Footer from "@/components/common/footer";
import Navbar from "@/components/common/navbar";
import { RootLayoutProvider } from "@/components/providers/RootLayoutProvider";

export const metadata: Metadata = {
  title: "Trackly",
  description: "Trackly",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body className="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden">
        <RootLayoutProvider>
          <Navbar />
          {children}
          <Footer />
        </RootLayoutProvider>
      </body>
    </html>
  );
}
