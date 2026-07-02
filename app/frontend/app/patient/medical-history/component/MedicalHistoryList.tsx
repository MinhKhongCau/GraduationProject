import { PlusCircle } from "lucide-react";
import { Button, Card } from "@/components/ui";
import type { MedicalHistory } from "@/types";

export interface MedicalHistoryListProps {
  histories: MedicalHistory[];
  onAdd: () => void;
}

export function MedicalHistoryList({ histories, onAdd }: MedicalHistoryListProps) {
  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-bold text-foreground">Medical History</h3>
        <Button size="sm" variant="soft" onClick={onAdd}>
          <PlusCircle className="h-4 w-4" /> Add entry
        </Button>
      </div>

      {histories.length === 0 ? (
        <p className="text-sm text-muted-foreground">No medical history recorded yet.</p>
      ) : (
        <div className="space-y-3">
          {histories.map((history) => (
            <Card key={history.historyId} className="p-4">
              <div className="mb-1 flex items-center justify-between">
                <p className="text-sm font-bold text-foreground">{history.conditionName}</p>
                {history.isChronic && (
                  <span className="rounded-md bg-warning-soft px-2 py-0.5 text-[10px] font-bold uppercase text-warning">
                    Chronic
                  </span>
                )}
              </div>
              {history.description && <p className="mb-1 text-sm text-muted-foreground">{history.description}</p>}
              <p className="text-xs text-muted-foreground">Diagnosed {history.diagnosedAt}</p>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
