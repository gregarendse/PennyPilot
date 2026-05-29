type AppShellProps = {
  children: React.ReactNode;
};

const navigationItems = ['Dashboard', 'Transactions', 'Budgets', 'Categories', 'Connections'];

export function AppShell({ children }: AppShellProps) {
  return (
    <main className="min-h-screen bg-slate-50 text-slate-950">
      <aside className="fixed inset-y-0 hidden w-64 border-r border-slate-200 bg-white p-6 lg:block">
        <p className="text-sm font-semibold uppercase tracking-[0.3em] text-mint">PennyPilot</p>
        <nav className="mt-10 space-y-2">
          {navigationItems.map((item) => (
            <a
              className="block rounded-xl px-4 py-3 text-sm font-medium text-slate-600 hover:bg-slate-100 hover:text-slate-950"
              href="#"
              key={item}
            >
              {item}
            </a>
          ))}
        </nav>
      </aside>
      <section className="lg:pl-64">{children}</section>
    </main>
  );
}
