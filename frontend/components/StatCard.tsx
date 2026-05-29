type StatCardProps = {
  label: string;
  value: string;
  helper: string;
};

export function StatCard({ label, value, helper }: StatCardProps) {
  return (
    <article className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <p className="text-sm font-medium text-slate-500">{label}</p>
      <p className="mt-3 text-3xl font-semibold tracking-tight text-slate-950">{value}</p>
      <p className="mt-2 text-sm text-slate-500">{helper}</p>
    </article>
  );
}
