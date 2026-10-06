import type { PageResult, PaginatedResponse } from "@/types";

/**
 * profile-service keeps personal info common to every role (full name, date of
 * birth, gender, phone number, country) in a nested `userInformation` object
 * on the base Profile, separate from the role-specific sub-profile.
 */
export interface RawUserInformation {
  fullName: string;
  dateOfBirth?: string | null;
  gender?: string;
  phoneNumber?: string;
  country?: string;
}

interface UserInformationFields {
  fullName: string;
  dateOfBirth?: string;
  gender?: string;
  phoneNumber?: string;
  country?: string;
}

/** Splits the app's flat profile request into profile-service's nested `userInformation`. */
export function splitUserInformation<T extends UserInformationFields>(
  payload: T,
) {
  const { fullName, dateOfBirth, gender, phoneNumber, country, ...rest } =
    payload;
  return {
    userInformation: { fullName, dateOfBirth, gender, phoneNumber, country },
    ...rest,
  };
}

/** Maps profile-service's PageResult into the app-wide PaginatedResponse shape. */
export function toPaginated<R, T>(
  page: PageResult<R>,
  map: (raw: R) => T,
): PaginatedResponse<T> {
  return {
    items: page.items.map(map),
    page: page.page,
    pageSize: page.pageSize,
    totalItems: page.total,
    totalPages: page.totalPages,
  };
}
