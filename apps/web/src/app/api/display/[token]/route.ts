import { NextRequest, NextResponse } from "next/server";
const api = process.env.INTERNAL_API_URL ?? "http://api:4000/api/v1";
export const dynamic = "force-dynamic";
export async function GET(_: NextRequest, context: { params: Promise<{ token: string }> }) {
  const { token } = await context.params;
  const upstream = await fetch(`${api}/display/${encodeURIComponent(token)}`, { cache: "no-store" });
  return new NextResponse(await upstream.text(), { status: upstream.status, headers: { "content-type": upstream.headers.get("content-type") ?? "application/json", "cache-control": "no-store" } });
}
