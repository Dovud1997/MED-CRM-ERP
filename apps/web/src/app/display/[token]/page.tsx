import { QueueTVDisplay } from "@/components/queue-tv-display";
export default async function DisplayPage({ params }: { params: Promise<{ token: string }> }) { const { token } = await params; return <QueueTVDisplay token={token}/>; }
