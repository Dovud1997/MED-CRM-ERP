import type { Metadata } from 'next'; import './globals.css';
export const metadata: Metadata={title:'ONA VA BOLA KLINIKASI',description:'Управление частной клиникой'};
export default function RootLayout({children}:{children:React.ReactNode}){return <html lang="ru"><body>{children}</body></html>}
