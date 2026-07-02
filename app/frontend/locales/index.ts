import enCommon from "./en/common.json";
import enAuth from "./en/auth.json";
import enPatient from "./en/patient.json";
import enExpert from "./en/expert.json";
import enAdmin from "./en/admin.json";
import enLanding from "./en/landing.json";

import viCommon from "./vi/common.json";
import viAuth from "./vi/auth.json";
import viPatient from "./vi/patient.json";
import viExpert from "./vi/expert.json";
import viAdmin from "./vi/admin.json";
import viLanding from "./vi/landing.json";

function deepMerge(target: Record<string, unknown>, source: Record<string, unknown>) {
  const result: Record<string, unknown> = { ...target };
  for (const [key, value] of Object.entries(source)) {
    const existing = result[key];
    if (
      value &&
      typeof value === "object" &&
      !Array.isArray(value) &&
      existing &&
      typeof existing === "object" &&
      !Array.isArray(existing)
    ) {
      result[key] = deepMerge(existing as Record<string, unknown>, value as Record<string, unknown>);
    } else {
      result[key] = value;
    }
  }
  return result;
}

export const LOCALE_DICTS = {
  en: [enCommon, enAuth, enPatient, enExpert, enAdmin, enLanding].reduce(
    (acc, dict) => deepMerge(acc, dict),
    {}
  ),
  vi: [viCommon, viAuth, viPatient, viExpert, viAdmin, viLanding].reduce(
    (acc, dict) => deepMerge(acc, dict),
    {}
  ),
};
