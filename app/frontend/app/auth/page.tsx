import { redirect } from "next/navigation";
import { ROUTES } from "@/constants";

export default function AuthIndexPage() {
  redirect(ROUTES.AUTH.LOGIN);
}
