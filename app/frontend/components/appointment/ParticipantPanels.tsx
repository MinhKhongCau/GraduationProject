"use client";

import { useId, useState, type ReactNode } from "react";
import { ChevronDown, PlayCircle } from "lucide-react";
import clsx from "clsx";
import { Badge, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { expertApi, patientApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import type { AppointmentParty } from "@/types";

const GENDER_TEXT: Record<string, string> = { MALE: "Nam", FEMALE: "Nữ", OTHER: "Khác" };
const ROLE_TEXT: Record<string, string> = { PATIENT: "Bệnh nhân", EXPERT: "Chuyên gia", ADMIN: "Quản trị viên" };
const VERIFICATION_TEXT: Record<string, { label: string; tone: "success" | "warning" | "neutral" | "danger" }> = {
  VERIFIED: { label: "Đã xác minh", tone: "success" },
  PENDING: { label: "Chờ duyệt", tone: "warning" },
  UNVERIFIED: { label: "Chưa xác minh", tone: "neutral" },
  REJECTED: { label: "Đã từ chối", tone: "danger" },
};

function DetailRow({ label, children }: { label: string; children?: ReactNode }) {
  return (
    <div className="grid grid-cols-3 gap-3 py-1.5 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="col-span-2 break-words text-foreground">{children || "—"}</dd>
    </div>
  );
}

function Avatar({ name, url }: { name?: string; url?: string }) {
  if (url) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={url} alt="" className="h-10 w-10 shrink-0 rounded-full object-cover" />;
  }
  return (
    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-sm font-semibold text-primary">
      {(name || "?").charAt(0).toUpperCase()}
    </div>
  );
}

interface ExpandablePartyProps {
  party?: AppointmentParty;
  fallbackId: string;
  /** Rendered only once expanded, so the detail request fires on first click. */
  renderDetail: () => ReactNode;
}

/** Header row (avatar, name, contact) that toggles a detail panel below it. */
function ExpandableParty({ party, fallbackId, renderDetail }: ExpandablePartyProps) {
  const [open, setOpen] = useState(false);
  const panelId = useId();

  return (
    <div>
      <button
        type="button"
        aria-expanded={open}
        aria-controls={panelId}
        onClick={() => setOpen((value) => !value)}
        className="flex w-full items-center gap-3 rounded-xl px-4 py-3 text-left transition-colors hover:bg-surface/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30"
      >
        <Avatar name={party?.fullName} url={party?.avatarUrl} />
        <div className="min-w-0 flex-1 text-sm">
          <p className="font-semibold text-foreground">{party?.fullName || `#${fallbackId.slice(0, 8)}`}</p>
          <p className="truncate text-xs text-muted-foreground">
            {[party?.email, party?.phoneNumber].filter(Boolean).join(" · ") || "—"}
          </p>
        </div>
        <span className="hidden text-xs font-medium text-primary sm:inline">{open ? "Thu gọn" : "Xem chi tiết"}</span>
        <ChevronDown
          aria-hidden="true"
          className={clsx("h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200", open && "rotate-180")}
        />
      </button>
      {open && (
        <div id={panelId} className="border-t border-border px-4 py-3">
          {renderDetail()}
        </div>
      )}
    </div>
  );
}

/** Public expert profile (GET /profiles/experts/:accountId): verification, specializations, bio. */
function ExpertDetail({ accountId, party }: { accountId: string; party?: AppointmentParty }) {
  const { data: expert, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.expertProfile(accountId),
    queryFn: () => expertApi.getExpertProfile(accountId),
  });
  if (isLoading) return <Spinner className="h-5 w-5" />;

  const verification = VERIFICATION_TEXT[expert?.verificationStatus ?? party?.verificationStatus ?? ""];
  return (
    <dl>
      <DetailRow label="Xác minh">
        {verification && <Badge tone={verification.tone}>{verification.label}</Badge>}
      </DetailRow>
      <DetailRow label="Chuyên khoa">
        {expert?.specializations.length ? (
          <div className="flex flex-wrap gap-1.5">
            {expert.specializations.map((spec) => (
              <Badge key={spec.specId} tone="primary">
                {spec.name}
              </Badge>
            ))}
          </div>
        ) : null}
      </DetailRow>
      <DetailRow label="Email">{expert?.email || party?.email}</DetailRow>
      <DetailRow label="Số điện thoại">{expert?.phoneNumber || party?.phoneNumber}</DetailRow>
      <DetailRow label="Giới tính">{expert?.gender ? GENDER_TEXT[expert.gender] ?? expert.gender : undefined}</DetailRow>
      <DetailRow label="Quốc gia">{expert?.country}</DetailRow>
      <DetailRow label="Giới thiệu">
        {expert?.bio && <p className="whitespace-pre-line text-sm leading-relaxed">{expert.bio}</p>}
      </DetailRow>
      {expert?.introductionVideoUrl && (
        <DetailRow label="Video">
          <a
            href={expert.introductionVideoUrl}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-1.5 font-semibold text-primary hover:underline"
          >
            <PlayCircle className="h-3.5 w-3.5" /> Xem video giới thiệu
          </a>
        </DetailRow>
      )}
    </dl>
  );
}

/**
 * Booking account. A patient sees their own full profile and an admin sees the patient profile;
 * an expert gets only what booking-service attached (profile-service hides patient data from experts).
 */
function BookingAccountDetail({ accountId, party }: { accountId: string; party?: AppointmentParty }) {
  const { user } = useAuthContext();
  const isSelf = user?.id === accountId;
  const canReadProfile = isSelf || user?.role === "ADMIN";

  const { data: profile, isLoading } = useApiQuery({
    queryKey: isSelf ? QUERY_KEYS.myPatientProfile() : QUERY_KEYS.patientProfile(accountId),
    queryFn: () => (isSelf ? patientApi.getMyProfile() : patientApi.getPatientProfile(accountId)),
    enabled: canReadProfile,
  });
  if (canReadProfile && isLoading) return <Spinner className="h-5 w-5" />;

  return (
    <dl>
      <DetailRow label="Vai trò">{party?.role ? ROLE_TEXT[party.role] ?? party.role : undefined}</DetailRow>
      <DetailRow label="Email">{profile?.email || party?.email}</DetailRow>
      <DetailRow label="Số điện thoại">{profile?.phoneNumber || party?.phoneNumber}</DetailRow>
      {canReadProfile && (
        <>
          <DetailRow label="Ngày sinh">
            {profile?.dateOfBirth ? new Date(profile.dateOfBirth).toLocaleDateString("vi-VN") : undefined}
          </DetailRow>
          <DetailRow label="Giới tính">
            {profile?.gender ? GENDER_TEXT[profile.gender] ?? profile.gender : undefined}
          </DetailRow>
          <DetailRow label="Địa chỉ">{profile?.address}</DetailRow>
          <DetailRow label="Quốc gia">{profile?.country}</DetailRow>
        </>
      )}
      {isSelf && <p className="mt-2 text-xs text-muted-foreground">Đây là tài khoản của bạn.</p>}
    </dl>
  );
}

export function ExpertPanel({ accountId, party }: { accountId: string; party?: AppointmentParty }) {
  return (
    <ExpandableParty
      party={party}
      fallbackId={accountId}
      renderDetail={() => <ExpertDetail accountId={accountId} party={party} />}
    />
  );
}

export function BookingAccountPanel({ accountId, party }: { accountId: string; party?: AppointmentParty }) {
  return (
    <ExpandableParty
      party={party}
      fallbackId={accountId}
      renderDetail={() => <BookingAccountDetail accountId={accountId} party={party} />}
    />
  );
}
