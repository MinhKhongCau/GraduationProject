"use client";

import { useState } from "react";
import Link from "next/link";
import { ClipboardList, Receipt, Wallet } from "lucide-react";
import { Card, PageHeader, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { expertApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";
import { OrdersTab } from "./component/OrdersTab";
import { CompensationCasesTab } from "./component/CompensationCasesTab";
import { WalletLedgerTab } from "./component/WalletLedgerTab";

type Tab = "orders" | "cases" | "ledger";

const TABS: { id: Tab; label: string; icon: typeof Receipt }[] = [
  { id: "orders", label: "Đơn thanh toán", icon: Receipt },
  { id: "cases", label: "Hồ sơ cần xử lý", icon: ClipboardList },
  { id: "ledger", label: "Sổ cái ví", icon: Wallet },
];

/** profile-service caps page_size at 100; an admin managing more would need paging here. */
const MANAGED_EXPERTS_PARAMS = { page: 1, pageSize: 100 };

/**
 * Admin transaction management. An admin manages the experts they approved, so every list
 * here is scoped server-side to those experts.
 */
export default function AdminTransactionsPage() {
  const [tab, setTab] = useState<Tab>("orders");

  const { data: managed, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.managedExperts(MANAGED_EXPERTS_PARAMS),
    queryFn: () => expertApi.listManagedExperts(MANAGED_EXPERTS_PARAMS),
  });
  const experts = managed?.items ?? [];
  const names = new Map(experts.map((expert) => [expert.accountId, expert.fullName || expert.email]));
  const expertName = (expertId: string) => names.get(expertId) ?? expertId.slice(0, 8);

  return (
    <div className="mx-auto max-w-7xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý giao dịch"
        description="Giao dịch của các chuyên gia bạn đã duyệt. Admin duyệt chuyên gia sẽ trở thành người quản lý chuyên gia đó."
      />

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : experts.length === 0 ? (
        <Card className="p-10 text-center">
          <p className="text-sm text-muted-foreground">
            Bạn chưa quản lý chuyên gia nào. Hãy duyệt hồ sơ chuyên gia để quản lý giao dịch của họ.
          </p>
          <Link href={ROUTES.ADMIN.EXPERTS} className="mt-3 inline-block text-sm font-semibold text-primary hover:underline">
            Đi tới Quản lý chuyên gia
          </Link>
        </Card>
      ) : (
        <>
          <div role="tablist" className="flex max-w-xl gap-1 rounded-lg border border-border bg-background p-1">
            {TABS.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                type="button"
                role="tab"
                aria-selected={tab === id}
                onClick={() => setTab(id)}
                className={`flex h-9 flex-1 items-center justify-center gap-2 rounded-md px-3 text-sm font-semibold transition-colors ${
                  tab === id ? "bg-background text-primary shadow-card" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                <Icon className="h-4 w-4" aria-hidden="true" />
                {label}
              </button>
            ))}
          </div>

          {tab === "orders" && <OrdersTab experts={experts} expertName={expertName} />}
          {tab === "cases" && <CompensationCasesTab experts={experts} expertName={expertName} />}
          {tab === "ledger" && <WalletLedgerTab experts={experts} expertName={expertName} />}
        </>
      )}
    </div>
  );
}
