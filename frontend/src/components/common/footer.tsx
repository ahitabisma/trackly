export default function Footer() {
  return (
    <footer
      style={{ borderTop: "1.5px solid rgba(26,22,18,.1)" }}
      className="bg-cream"
    >
      <div className="max-w-7xl mx-auto px-8 py-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <a href="#" className="font-serif text-lg text-ink">
            Trackly
          </a>
          <p className="text-xs text-muted mt-1">
            Data sourced from KSEI. For research &amp; educational purposes
            only.
          </p>
        </div>
        <div className="flex gap-6 text-sm text-muted">
          <a href="#" className="hover:text-ink transition-colors">
            About
          </a>
          <a href="#" className="hover:text-ink transition-colors">
            API Docs
          </a>
          <a href="#" className="hover:text-ink transition-colors">
            GitHub
          </a>
        </div>
        <div className="font-mono text-xs text-muted">© {new Date().getFullYear()} Trackly</div>
      </div>
    </footer>
  );
}
