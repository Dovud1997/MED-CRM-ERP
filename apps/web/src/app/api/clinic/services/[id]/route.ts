import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export async function PATCH(request: NextRequest, context: { params: Promise<{ id: string }> }) { const { id } = await context.params; return clinicResponse(request, `/services/${encodeURIComponent(id)}`, { method: "PATCH", headers: { "content-type": "application/json" }, body: await request.text() }); }
export async function DELETE(request: NextRequest, context: { params: Promise<{ id: string }> }) { const { id } = await context.params; return clinicResponse(request, `/services/${encodeURIComponent(id)}`, { method: "DELETE" }); }
