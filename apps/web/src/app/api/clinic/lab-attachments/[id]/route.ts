import {NextRequest,NextResponse} from "next/server";
import {clinicRequest} from "@/lib/server-api";
export async function GET(request:NextRequest,{params}:{params:Promise<{id:string}>}){
  const{id}=await params;const{upstream,rotated}=await clinicRequest(request,`/lab-attachments/${encodeURIComponent(id)}`);
  const response=new NextResponse(upstream.body,{status:upstream.status,headers:{"content-type":upstream.headers.get("content-type")??"application/octet-stream","content-disposition":upstream.headers.get("content-disposition")??"inline","cache-control":"private, no-store"}});
  if(rotated){const secure=process.env.COOKIE_SECURE==="true";response.cookies.set("obk_access",rotated.accessToken,{httpOnly:true,sameSite:"strict",secure,path:"/",maxAge:900});response.cookies.set("obk_refresh",rotated.refreshToken,{httpOnly:true,sameSite:"strict",secure,path:"/api/session",maxAge:2592000});}
  return response;
}
