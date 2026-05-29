import { AppShell } from '@/components/AppShell';
import { StatCard } from '@/components/StatCard';

const nextSteps = [
  'Connect Monzo with the direct OAuth flow',
  'Add TrueLayer for Barclays and Barclaycard',
  'Upload American Express CSV exports',
  'Create budget categories and monthly targets',
];

export default function Home() {
  return (
    <AppShell>
      <div className="mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10">
        <header className="rounded-3xl bg-ink p-8 text-white shadow-sm">
          <p className="text-sm font-semibold uppercase tracking-[0.3em] text-mint">Bootstrap preview</p>
          <h1 className="mt-4 max-w-3xl text-4xl font-bold tracking-tight">Your self-hosted command centre for everyday spending.</h1>
          <p className="mt-4 max-w-2xl text-slate-300">
            This placeholder dashboard establishes the app shell while the bank sync, categorisation, and budget views are implemented.
          </p>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          <StatCard label="Connected accounts" value="0" helper="Monzo, Barclays, and CSV import connectors are scaffolded." />
          <StatCard label="Transactions" value="0" helper="The API is ready for idempotent transaction ingestion." />
          <StatCard label="Budget coverage" value="0%" helper="Monthly category budgets will populate this view." />
        </section>

        <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-xl font-semibold text-slate-950">Implementation roadmap</h2>
          <ul className="mt-4 grid gap-3 md:grid-cols-2">
            {nextSteps.map((step) => (
              <li className="rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600" key={step}>
                {step}
              </li>
            ))}
          </ul>
        </section>
      </div>
    </AppShell>
  );
}
