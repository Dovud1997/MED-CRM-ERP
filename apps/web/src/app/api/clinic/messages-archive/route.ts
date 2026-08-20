import {NextRequest} from "next/server";
import {clinicResponse} from "@/lib/server-api";
export const dynamic="force-dynamic";
export async function GET(request:NextRequest){return clinicResponse(request,`/messages-archive${request.nextUrl.search}`)}
