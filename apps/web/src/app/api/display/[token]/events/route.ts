const api = process.env.INTERNAL_API_URL ?? "http://api:4000/api/v1";
export const dynamic = "force-dynamic";
export async function GET(_: Request, context: { params: Promise<{ token: string }> }) {
  const { token } = await context.params;
  const upstream = await fetch(`${api}/display/${encodeURIComponent(token)}/events`, { cache: "no-store" });
  return new Response(upstream.body, { status: upstream.status, headers: { "content-type": upstream.headers.get("content-type") ?? "text/event-stream", "cache-control": "no-cache", connection: "keep-alive" } });
}
