import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";

export const dynamic = "force-dynamic";

async function forward(request: NextRequest, method: string, parts?: string[]) {
  const suffix = parts?.length ? `/${parts.join("/")}` : "";
  const contentType = request.headers.get("content-type") ?? "";
  const init: RequestInit & { duplex?: "half" } = { method };
  if (method !== "GET" && method !== "HEAD") {
    init.body = request.body;
    init.duplex = "half";
  }
  if (contentType) init.headers = { "content-type": contentType };
  return clinicResponse(request, `/queue${suffix}${request.nextUrl.search}`, init);
}

export async function GET(request: NextRequest, context: { params: Promise<{ path?: string[] }> }) { return forward(request, "GET", (await context.params).path); }
export async function POST(request: NextRequest, context: { params: Promise<{ path?: string[] }> }) { return forward(request, "POST", (await context.params).path); }
export async function PATCH(request: NextRequest, context: { params: Promise<{ path?: string[] }> }) { return forward(request, "PATCH", (await context.params).path); }
export async function DELETE(request: NextRequest, context: { params: Promise<{ path?: string[] }> }) { return forward(request, "DELETE", (await context.params).path); }
