import { AlertTriangle } from "lucide-react";
import { Card, Button } from "@/components/ui";

/** No backend endpoint exists for self-service account deletion — disabled. */
export function DeleteAccountSection() {
  return (
    <Card className="border-danger/20 p-6">
      <div className="mb-3 flex items-center gap-2 text-danger">
        <AlertTriangle className="h-4 w-4" />
        <h3 className="text-sm font-bold">Delete Account</h3>
      </div>
      <p className="mb-4 text-sm text-muted-foreground">
        Permanently delete your account and all associated data. This action cannot be undone.
      </p>
      <Button variant="danger" disabled title="Account deletion isn't available yet — contact support.">
        Delete my account
      </Button>
    </Card>
  );
}
