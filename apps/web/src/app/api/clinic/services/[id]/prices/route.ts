import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export async function POST(request: NextRequest, context: { params: Promise<{ id: string }> }) { const { id } = await context.params; return clinicResponse(request, `/services/${encodeURIComponent(id)}/prices`, { method: "POST", headers: { "content-type": "application/json" }, body: await request.text() }); }
