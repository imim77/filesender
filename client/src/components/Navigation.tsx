type NavigationProps = {
  logoSrc?: string;
  logoAlt?: string;
  aboutHref?: string;
  className?: string;
};

export default function Navigation({
  logoSrc = '',
  logoAlt = 'Filesender',
  aboutHref = '#',
  className = '',
}: NavigationProps) {
  return (
    <header data-slot="navigation" className={`shrink-0 border-b bg-background ${className}`.trim()}>
      <nav className="flex w-full items-center justify-between gap-4 px-4 py-3 md:px-6">
        <div className="flex min-w-0 items-center gap-3">
          {logoSrc ? (
            <a
              href="https://github.com/imim77/filesender"
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-sm outline-none transition-colors focus-visible:ring-3 focus-visible:ring-ring/50"
            >
              <img src={logoSrc} alt={logoAlt} className="h-8" />
            </a>
          ) : null}
          <span className="truncate text-sm font-semibold">Filesender</span>
        </div>
          <a
            href={aboutHref}
          className="shrink-0 rounded-sm text-sm font-medium text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            About project
          </a>
      </nav>
    </header>
  );
}
