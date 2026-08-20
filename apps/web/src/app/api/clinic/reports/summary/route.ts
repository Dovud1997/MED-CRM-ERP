import { NextRequest } from "next/server"; import { clinicResponse } from "@/lib/server-api"; export const dynamic="force-dynamic";
export async function GET(r:NextRequest){return clinicResponse(r,`/reports/summary${r.nextUrl.search}`)}
