import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export async function GET(request: NextRequest) { return clinicResponse(request, "/services"); }
export async function POST(request: NextRequest) { return clinicResponse(request, "/services", { method: "POST", headers: { "content-type": "application/json" }, body: await request.text() }); }
