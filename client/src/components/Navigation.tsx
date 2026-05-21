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
    <header data-slot="navigation" className={`sticky top-0 z-50 w-full bg-background ${className}`.trim()}>
      <nav className="flex w-full items-center justify-start px-6 py-4">
        <div className="flex items-center gap-6">
          {logoSrc ? (
            <a
              href="https://github.com/imim77/filesender"
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-sm outline-none transition-colors focus-visible:ring-3 focus-visible:ring-ring/50"
            >
              <img src={logoSrc} alt={logoAlt} className="h-12" />
            </a>
          ) : null}
          <a
            href={aboutHref}
            className="rounded-sm text-base font-medium text-foreground outline-none transition-colors hover:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            About project
          </a>
        </div>
      </nav>
    </header>
  );
}
