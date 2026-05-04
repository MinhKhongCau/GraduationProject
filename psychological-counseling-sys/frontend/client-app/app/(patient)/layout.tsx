'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  LayoutDashboard, 
  CalendarPlus, 
  Wallet, 
  ClipboardList, 
  MessageSquare, 
  Settings
} from 'lucide-react';

export default function PatientLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();

  const menuItems = [
    { name: 'Dashboard', icon: LayoutDashboard, href: '/dashboard' },
    { name: 'Book Appointment', icon: CalendarPlus, href: '/book-appointment' },
    { name: 'Wallet', icon: Wallet, href: '/wallet' },
    { name: 'Assessment', icon: ClipboardList, href: '/assessment' },
    { name: 'Messages', icon: MessageSquare, href: '/messages' },
  ];

  return (
    /* KHÓA CHIỀU CAO: Dùng h-screen và overflow-hidden ở lớp ngoài cùng 
       để chặn việc cuộn toàn bộ trang web.
    */
    <div className="h-screen bg-white flex flex-col font-sans text-slate-800 overflow-hidden">
      
      {/* FIXED HEADER: Thêm sticky, top-0 và z-index cao 
      */}
      <header className="h-16 border-b border-slate-200 flex items-center justify-between px-6 bg-white shrink-0 z-30 sticky top-0">
        <div className="flex items-center gap-2 text-xl font-bold">
          <span className="bg-sky-400 text-white p-1 rounded-lg">🧠</span>
          MindCare
        </div>
        <div className="flex items-center gap-4 text-sm font-medium">
          <span className="text-slate-500">
            Balance: <span className="text-slate-900 font-bold">$250.00</span>
          </span>
          <div className="w-8 h-8 bg-slate-100 border border-slate-200 rounded-full flex items-center justify-center font-bold text-slate-600">
            SD
          </div>
        </div>
      </header>

      {/* THÂN MÁY (Sidebar + Content) */}
      <div className="flex flex-1 overflow-hidden">
        
        {/* FIXED SIDEBAR: Nhờ overflow-hidden ở cha, Sidebar sẽ đứng yên.
           Chúng ta có thể thêm overflow-y-auto cho chính nó nếu menu quá dài.
        */}
        <aside className="w-64 border-r border-slate-200 bg-white flex flex-col justify-between shrink-0 hidden md:flex h-full">
          <nav className="p-4 space-y-1 mt-2">
            {menuItems.map((item) => {
              const isActive = pathname === item.href;
              const Icon = item.icon;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={`flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                    isActive 
                      ? 'bg-sky-400 text-white shadow-sm' 
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  {item.name}
                </Link>
              );
            })}
          </nav>
          
          <div className="p-4 border-t border-slate-100 mt-auto">
            <Link
              href="/settings"
              className="flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium text-slate-600 hover:bg-slate-50 hover:text-slate-900 transition-colors"
            >
              <Settings className="w-4 h-4" />
              Settings
            </Link>
          </div>
        </aside>

        {/* SCROLLABLE AREA: Đây là phần duy nhất được phép cuộn.
           Phần này chứa trang nội dung (Dashboard, Wallet...) và Footer.
        */}
        <div className="flex-1 flex flex-col overflow-y-auto bg-slate-50/30 scroll-smooth">
          
          <main className="flex-1 p-8">
            {children}
          </main>

          {/* FOOTER nằm cuối vùng cuộn */}
          <footer className="bg-slate-50 border-t border-slate-200 px-8 py-10">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
              <div className="col-span-1">
                <div className="flex items-center gap-2 text-xl font-bold mb-3">
                  <span className="text-sky-400">🧠</span>
                  MindCare
                </div>
                <p className="text-xs text-slate-500 leading-relaxed pr-4">
                  Empowering mental well-being through compassionate care.
                </p>
              </div>
              <div>
                <h4 className="font-bold text-sm mb-4">Services</h4>
                <ul className="space-y-2 text-xs text-slate-500">
                  <li><a href="#" className="hover:text-sky-500 transition-colors">Book an Appointment</a></li>
                  <li><a href="#" className="hover:text-sky-500 transition-colors">Psychological Assessment</a></li>
                  <li><a href="#" className="hover:text-sky-500 transition-colors">Find an Expert</a></li>
                </ul>
              </div>
              <div>
                <h4 className="font-bold text-sm mb-4">Company</h4>
                <ul className="space-y-2 text-xs text-slate-500">
                  <li><a href="#" className="hover:text-sky-500 transition-colors">About Us</a></li>
                  <li><a href="#" className="hover:text-sky-500 transition-colors">Careers</a></li>
                  <li><a href="#" className="hover:text-sky-500 transition-colors">Privacy Policy</a></li>
                  <li><a href="#" className="hover:text-sky-500 transition-colors">Terms of Service</a></li>
                </ul>
              </div>
              <div>
                <h4 className="font-bold text-sm mb-4">Connect</h4>
                <div className="flex gap-4 text-slate-400">
                  {/* Mã SVG Facebook, Twitter, Linkedin bạn đã cập nhật ở đây */}
                </div>
              </div>
            </div>
            <div className="text-center text-xs text-slate-400 pt-6 border-t border-slate-200">
              © 2026 MindCare. All rights reserved.
            </div>
          </footer>

        </div>
      </div>
    </div>
  );
}