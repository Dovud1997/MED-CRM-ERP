import{NextRequest}from"next/server";import{clinicResponse}from"@/lib/server-api";
export async function POST(r:NextRequest){return clinicResponse(r,"/cash/shifts",{method:"POST",headers:{"content-type":"application/json"},body:await r.text()})}
