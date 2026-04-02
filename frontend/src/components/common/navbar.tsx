"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

export default function Navbar() {
  const pathname = usePathname();

  const getLinkClassName = (href: string) => {
    let classes = "neo-btn-sm";

    if (pathname === href) {
      classes += " active";
    }

    return classes;
  };

  return (
    <nav
      className="fixed top-0 left-0 right-0 z-50 bg-cream"
      style={{ borderBottom: "1.5px solid rgba(26,22,18,.1)" }}
    >
      <div className="max-w-7xl mx-auto px-5 sm:px-8 h-14 flex items-center justify-between">
        <Link href="/" className="font-serif text-xl text-ink tracking-tight">
          TRACKLY
        </Link>
        <div className="flex items-center gap-5">
          <Link href="/" className={getLinkClassName("/")}>
            1% Share Ownership
          </Link>
          <Link
            href="/shareholder-concentration-list"
            className={getLinkClassName("/shareholder-concentration-list")}
          >
            Shareholder Concentration List
          </Link>
        </div>
      </div>
    </nav>
  );
}
