export function ProgressBar({ current, total }: { current: number; total: number }) {
  const percentage = total === 0 ? 0 : Math.round((current / total) * 100);

  return (
    <div>
      <div className="mb-2 flex justify-between text-xs font-semibold text-muted-foreground">
        <span>
          Question {Math.min(current + 1, total)} of {total}
        </span>
        <span>{percentage}%</span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-surface">
        <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${percentage}%` }} />
      </div>
    </div>
  );
}
