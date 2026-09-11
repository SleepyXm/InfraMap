import type { Metadata } from "next";
import "./globals.css";
import Navbar from "./components/InfraMapNav";
import { UserProvider } from "@/app/components/provider/UserProvider";
import { Geist, Geist_Mono } from "next/font/google";
import "@/app/globals.css";
import { InteractiveGrid } from "@/app/UI";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});


export const metadata: Metadata = {
  title: "InfraMap",
  description: "A live operational map for your infrastructure.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable} ${geistSans.className} antialiased`}>
        <InteractiveGrid>
          <div className="relative isolate min-h-screen w-full">
            <UserProvider>
              <Navbar />
              {children}
            </UserProvider>
          </div>
        </InteractiveGrid>
      </body>
    </html>
  );
}
