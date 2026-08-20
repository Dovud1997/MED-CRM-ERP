import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export const dynamic = "force-dynamic";
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ employeeId: string }> },
) {
  const { employeeId } = await params;
  return clinicResponse(request, `/doctor-schedules/${employeeId}`);
}
export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ employeeId: string }> },
) {
  const { employeeId } = await params;
  return clinicResponse(request, `/doctor-schedules/${employeeId}`, {
    method: "PUT",
    body: await request.text(),
  });
}
