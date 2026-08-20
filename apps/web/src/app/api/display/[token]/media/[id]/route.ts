const api = process.env.INTERNAL_API_URL ?? "http://api:4000/api/v1";
export const dynamic = "force-dynamic";
export async function GET(_: Request, context: { params: Promise<{ token: string; id: string }> }) {
  const { token, id } = await context.params;
  const upstream = await fetch(`${api}/display/${encodeURIComponent(token)}/media/${encodeURIComponent(id)}`, { cache: "no-store" });
  return new Response(upstream.body, { status: upstream.status, headers: { "content-type": upstream.headers.get("content-type") ?? "application/octet-stream", "cache-control": upstream.headers.get("cache-control") ?? "private, max-age=3600" } });
}
