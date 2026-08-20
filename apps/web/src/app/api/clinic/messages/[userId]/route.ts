import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export const dynamic = "force-dynamic";
export async function GET(request: NextRequest, context: { params: Promise<{ userId: string }> }) { const { userId } = await context.params; return clinicResponse(request, `/messages/${userId}`); }
export async function POST(request: NextRequest, context: { params: Promise<{ userId: string }> }) { const { userId } = await context.params; return clinicResponse(request, `/messages/${userId}`, { method: "POST", body: await request.text() }); }
