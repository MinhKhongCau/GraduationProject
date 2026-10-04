import { PlusCircle } from "lucide-react";
import { Badge, Button, Card } from "@/components/ui";
import type { MedicalHistory } from "@/types";

export interface MedicalHistoryListProps {
  histories: MedicalHistory[];
  onAdd: () => void;
}

export function MedicalHistoryList({ histories, onAdd }: MedicalHistoryListProps) {
  return (
    <div>
      <div className="mb-3 flex items-center justify-between gap-3">
        <h2 className="text-base font-semibold text-foreground">Medical History</h2>
        <Button size="sm" variant="soft" onClick={onAdd}>
          <PlusCircle className="h-4 w-4" /> Add entry
        </Button>
      </div>

      {histories.length === 0 ? (
        <Card className="p-8 text-center text-sm text-muted-foreground">No medical history recorded yet.</Card>
      ) : (
        <div className="space-y-3">
          {histories.map((history) => (
            <Card key={history.historyId} className="p-4">
              <div className="mb-1 flex items-center justify-between gap-3">
                <p className="text-sm font-semibold text-foreground">{history.conditionName}</p>
                {history.isChronic && (
                  <Badge tone="warning">Chronic</Badge>
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
