'use client';

import React from 'react';
import { 
  Wallet, 
  PlusCircle, 
  ArrowDownToLine, 
  Send, 
  Download, 
  BarChart2, 
  ArrowDownLeft, 
  ArrowUpRight, 
  Building, 
  RotateCcw, 
  Info, 
  CreditCard,
  ChevronRight
} from 'lucide-react';

// Dữ liệu giả lập cho Lịch sử giao dịch
const transactions = [
  { 
    id: 1, 
    title: 'Nạp tiền từ ngân hàng', 
    date: 'Hôm nay, 14:20', 
    amount: '+2.000.000 đ', 
    type: 'deposit', 
    status: 'Hoàn tất' 
  },
  { 
    id: 2, 
    title: 'Tư vấn Chuyên gia: BS. Lê Anh', 
    date: 'Hôm nay, 09:15', 
    amount: '-500.000 đ', 
    type: 'payment', 
    status: 'Hoàn tất' 
  },
  { 
    id: 3, 
    title: 'Rút tiền về Vietcombank', 
    date: 'Hôm qua, 18:30', 
    amount: '-1.000.000 đ', 
    type: 'withdraw', 
    status: 'Đang xử lý' 
  },
  { 
    id: 4, 
    title: 'Hoàn tiền: Gói đánh giá stress', 
    date: '12 Th04, 2024', 
    amount: '+250.000 đ', 
    type: 'refund', 
    status: 'Hoàn tất' 
  },
  { 
    id: 5, 
    title: 'Gia hạn gói Member Prime', 
    date: '10 Th04, 2024', 
    amount: '-199.000 đ', 
    type: 'payment', 
    status: 'Hoàn tất' 
  },
];

// Hàm phụ trợ để chọn Icon và Màu sắc dựa trên loại giao dịch
const getTransactionStyle = (type: string) => {
  switch (type) {
    case 'deposit': return { icon: ArrowDownLeft, bg: 'bg-cyan-50', text: 'text-cyan-500', amountColor: 'text-cyan-500' };
    case 'payment': return { icon: ArrowUpRight, bg: 'bg-rose-50', text: 'text-rose-500', amountColor: 'text-rose-500' };
    case 'withdraw': return { icon: Building, bg: 'bg-slate-100', text: 'text-slate-500', amountColor: 'text-rose-500' };
    case 'refund': return { icon: RotateCcw, bg: 'bg-slate-100', text: 'text-slate-600', amountColor: 'text-cyan-500' };
    default: return { icon: ArrowUpRight, bg: 'bg-slate-100', text: 'text-slate-500', amountColor: 'text-slate-900' };
  }
};

export default function WalletPage() {
  return (
    <div className="max-w-6xl mx-auto h-full">
      
      {/* Sử dụng CSS Grid để chia layout: 2 cột lớn bên trái, 1 cột nhỏ bên phải */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* CỘT TRÁI (Main Content) - Chiếm 2 phần */}
        <div className="lg:col-span-2 space-y-6">
          
          {/* 1. Thẻ Số dư chính */}
          <div className="bg-[#f0f9fa] rounded-3xl p-8 flex flex-col items-center justify-center text-center">
            <div className="w-12 h-12 bg-cyan-100 text-cyan-500 rounded-full flex items-center justify-center mb-4">
              <Wallet className="w-6 h-6" />
            </div>
            <p className="text-sm font-semibold text-slate-500 tracking-wider uppercase mb-2">Tổng số dư khả dụng</p>
            <h1 className="text-5xl font-extrabold text-slate-900 mb-8">
              15.450.000 <span className="text-3xl">đ</span>
            </h1>
            
            <div className="flex w-full max-w-md gap-4">
              <button className="flex-1 flex items-center justify-center gap-2 bg-[#6ee7b7] hover:bg-[#5eead4] text-slate-900 font-bold py-3 rounded-xl transition-colors">
                <PlusCircle className="w-5 h-5" /> Nạp tiền
              </button>
              <button className="flex-1 flex items-center justify-center gap-2 bg-white border border-cyan-200 text-cyan-600 hover:bg-cyan-50 font-bold py-3 rounded-xl transition-colors">
                Rút tiền
              </button>
            </div>
          </div>

          {/* 3. Lịch sử giao dịch */}
          <div className="bg-white pt-6">
            <div className="flex items-center justify-between mb-4 px-2">
              <h2 className="text-lg font-bold text-slate-900">Lịch sử giao dịch</h2>
              <button className="text-sm font-semibold text-cyan-500 flex items-center hover:text-cyan-600">
                Xem tất cả <ChevronRight className="w-4 h-4 ml-1" />
              </button>
            </div>
            
            <div className="space-y-0">
              {transactions.map((tx, index) => {
                const style = getTransactionStyle(tx.type);
                const Icon = style.icon;
                
                return (
                  <div key={tx.id} className={`flex items-center justify-between p-4 hover:bg-slate-50 transition-colors rounded-2xl ${index !== transactions.length - 1 ? 'border-b border-slate-100' : ''}`}>
                    <div className="flex items-center gap-4">
                      <div className={`w-10 h-10 rounded-full flex items-center justify-center ${style.bg} ${style.text}`}>
                        <Icon className="w-5 h-5" />
                      </div>
                      <div>
                        <p className="text-sm font-bold text-slate-900">{tx.title}</p>
                        <p className="text-xs font-medium text-slate-500 mt-0.5">{tx.date}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className={`text-sm font-bold ${style.amountColor}`}>{tx.amount}</p>
                      <span className={`inline-block mt-1 text-[10px] font-bold px-2 py-0.5 rounded-md ${
                        tx.status === 'Đang xử lý' ? 'bg-amber-50 text-amber-600' : 'bg-slate-100 text-slate-600'
                      }`}>
                        {tx.status}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

        </div>

        {/* CỘT PHẢI (Sidebar Info) - Chiếm 1 phần */}
        <div className="lg:col-span-1 space-y-6">
          
          {/* Thẻ Thu nhập Chuyên gia */}
          <div className="bg-white border-2 border-dashed border-cyan-100 rounded-3xl p-6 relative overflow-hidden">
            <span className="inline-block bg-cyan-400 text-white text-[10px] font-bold px-3 py-1 rounded-full uppercase tracking-wider mb-4">
              Dành cho Chuyên gia
            </span>
            <h3 className="text-base font-bold text-slate-900 mb-1">Thu nhập của bạn</h3>
            <p className="text-xs text-slate-500 mb-4">Tổng kết thu nhập từ các buổi tư vấn trực tuyến.</p>
            
            <div className="border border-slate-100 rounded-xl p-4 mb-4 bg-slate-50/50">
              <p className="text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1">Thu nhập tích lũy</p>
              <p className="text-2xl font-bold text-slate-900">12.450.000 đ</p>
            </div>
            
            <button className="w-full flex items-center justify-center gap-2 bg-white border border-cyan-200 text-cyan-600 hover:bg-cyan-50 font-bold py-2.5 rounded-xl text-sm transition-colors">
              <ArrowDownToLine className="w-4 h-4" /> Rút thu nhập chuyên gia
            </button>
            <p className="text-[10px] text-center text-slate-400 mt-3 font-medium">Giao dịch rút tiền sẽ được xử lý trong vòng 24h làm việc.</p>
          </div>

          {/* Thẻ Ưu đãi */}
          <div className="bg-gradient-to-br from-[#dcfce7] to-[#ccfbf1] rounded-3xl p-6 relative overflow-hidden">
            {/* Vòng tròn trang trí background */}
            <div className="absolute top-4 right-4 w-16 h-16 bg-white/20 rounded-full blur-xl"></div>
            
            <h3 className="text-base font-bold text-slate-900 mb-2 relative z-10">Ưu đãi gói hội viên</h3>
            <p className="text-xs text-slate-700 mb-5 relative z-10 leading-relaxed">
              Nhận hoàn tiền 5% cho tất cả dịch vụ đánh giá tâm lý khi thanh toán qua ví MindCare.
            </p>
            <button className="w-full bg-[#6ee7b7] hover:bg-[#5eead4] text-slate-900 font-bold py-2.5 rounded-xl text-sm transition-colors relative z-10">
              Tìm hiểu thêm
            </button>
          </div>

          {/* Thẻ Trợ giúp */}
          <div className="bg-white border border-slate-100 rounded-3xl p-6">
            <h3 className="text-sm font-bold text-slate-900 mb-4">Bạn cần giúp đỡ?</h3>
            <div className="space-y-3">
              <button className="w-full flex items-center gap-3 p-2 hover:bg-slate-50 rounded-lg transition-colors group">
                <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center group-hover:bg-cyan-50 transition-colors">
                  <Info className="w-4 h-4 text-slate-600 group-hover:text-cyan-500" />
                </div>
                <span className="text-sm font-semibold text-slate-700">Hướng dẫn nạp/rút tiền</span>
              </button>
              <button className="w-full flex items-center gap-3 p-2 hover:bg-slate-50 rounded-lg transition-colors group">
                <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center group-hover:bg-cyan-50 transition-colors">
                  <CreditCard className="w-4 h-4 text-slate-600 group-hover:text-cyan-500" />
                </div>
                <span className="text-sm font-semibold text-slate-700">Biểu phí giao dịch</span>
              </button>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}