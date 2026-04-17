'use client';

import React, { useState } from 'react';
import { 
  Search, 
  Phone, 
  Video, 
  MoreVertical, 
  Paperclip, 
  SendHorizonal, 
  CheckCheck,
  Circle
} from 'lucide-react';

// --- MOCK DATA ---
const contacts = [
  {
    id: 1,
    name: 'MSc. Nguyen An Binh',
    avatar: 'https://i.pravatar.cc/150?u=binh',
    lastMessage: 'Hello, I have received your information...',
    time: '10:30',
    unread: true,
    online: true,
  },
  {
    id: 2,
    name: 'Dr. Tran Minh Tam',
    avatar: 'https://i.pravatar.cc/150?u=tam',
    lastMessage: 'Don\'t forget our regular consultation schedule...',
    time: '08:15',
    unread: false,
    online: false,
  },
  {
    id: 3,
    name: 'MSc. Le Thi Hanh',
    avatar: 'https://i.pravatar.cc/150?u=hanh',
    lastMessage: 'Thank you for sharing your story with me.',
    time: 'Yesterday',
    unread: false,
    online: true,
  },
  {
    id: 4,
    name: 'PhD. Pham Quang Vinh',
    avatar: 'https://i.pravatar.cc/150?u=vinh',
    lastMessage: 'Please try the deep breathing exercises I...',
    time: 'Mon',
    unread: false,
    online: false,
  },
  {
    id: 5,
    name: 'Dr. Phan Hoang Long',
    avatar: 'https://i.pravatar.cc/150?u=long',
    lastMessage: 'Your new prescription has been sent to your...',
    time: 'Sat',
    unread: false,
    online: true,
  },
];

const chatHistory = [
  {
    id: 1,
    sender: 'patient',
    text: 'Hello Doctor, I just completed the psychological assessment and feel a bit anxious about the results.',
    time: '10:20',
  },
  {
    id: 2,
    sender: 'expert',
    text: 'Hello, I have received your information. Please don\'t worry too much, the assessment results are just the first step for us to better understand your current condition.',
    time: '10:22',
  },
  {
    id: 3,
    sender: 'expert',
    text: 'Your anxiety level is at a Moderate level. Do you frequently have trouble sleeping?',
    time: '10:23',
  },
  {
    id: 4,
    sender: 'patient',
    text: 'Yes, lately I\'ve been tossing and turning and overthinking about work before falling asleep.',
    time: '10:25',
  },
  {
    id: 5,
    sender: 'expert',
    text: 'Your assessment results show a slightly high anxiety level. I recommend trying some relaxation techniques before bed. We can start with the 4-7-8 breathing exercise.',
    time: '10:30',
  },
];

export default function MessagesPage() {
  const [activeContact, setActiveContact] = useState(contacts[0]);

  return (
    // Card container full chiều cao của phần content
    <div className="bg-white border border-slate-200 rounded-2xl flex overflow-hidden h-[750px] shadow-sm">
      
      {/* CỘT TRÁI: DANH SÁCH LIÊN HỆ */}
      <div className="w-80 flex-shrink-0 border-r border-slate-200 flex flex-col bg-white">
        
        {/* Header Left */}
        <div className="p-4 border-b border-slate-100">
          <h2 className="text-xl font-bold text-slate-900 mb-4">Messages</h2>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="h-4 w-4 text-slate-400" />
            </div>
            <input
              type="text"
              className="block w-full pl-9 pr-3 py-2 bg-slate-50 border-none rounded-xl text-sm focus:ring-2 focus:ring-sky-100 outline-none"
              placeholder="Search experts..."
            />
          </div>
        </div>

        {/* Contact List (Scrollable) */}
        <div className="flex-1 overflow-y-auto">
          {contacts.map((contact) => (
            <button
              key={contact.id}
              onClick={() => setActiveContact(contact)}
              className={`w-full flex items-center gap-3 p-4 text-left transition-colors border-b border-slate-50/50 ${
                activeContact.id === contact.id 
                  ? 'bg-sky-50/50 border-r-2 border-r-sky-500' 
                  : 'hover:bg-slate-50'
              }`}
            >
              {/* Avatar */}
              <div className="relative flex-shrink-0">
                <img 
                  src={contact.avatar} 
                  alt={contact.name} 
                  className="w-12 h-12 rounded-full object-cover border border-slate-200"
                />
                {contact.online && (
                  <span className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></span>
                )}
              </div>
              
              {/* Info */}
              <div className="flex-1 min-w-0">
                <div className="flex justify-between items-baseline mb-0.5">
                  <h3 className={`text-sm truncate ${contact.unread ? 'font-bold text-slate-900' : 'font-semibold text-slate-700'}`}>
                    {contact.name}
                  </h3>
                  <span className="text-xs text-slate-400 flex-shrink-0 ml-2">{contact.time}</span>
                </div>
                <div className="flex justify-between items-center">
                  <p className={`text-xs truncate ${contact.unread ? 'font-semibold text-slate-700' : 'text-slate-500'}`}>
                    {contact.lastMessage}
                  </p>
                  {contact.unread && (
                    <span className="w-2 h-2 bg-sky-500 rounded-full flex-shrink-0 ml-2"></span>
                  )}
                </div>
              </div>
            </button>
          ))}
        </div>
      </div>


      {/* CỘT PHẢI: KHUNG CHAT */}
      <div className="flex-1 flex flex-col bg-white min-w-0">
        
        {/* Chat Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-100 bg-white">
          <div className="flex items-center gap-3">
            <div className="relative">
              <img src={activeContact.avatar} alt="Avatar" className="w-10 h-10 rounded-full object-cover" />
              {activeContact.online && (
                <span className="absolute bottom-0 right-0 w-2.5 h-2.5 bg-green-500 border-2 border-white rounded-full"></span>
              )}
            </div>
            <div>
              <h3 className="font-bold text-slate-900 text-sm">{activeContact.name}</h3>
              <div className="flex items-center gap-1.5 text-xs text-slate-500">
                <span>Clinical Psychologist</span>
                {activeContact.online && (
                  <>
                    <Circle className="w-1 h-1 fill-slate-300 text-slate-300" />
                    <span className="text-cyan-500 font-medium">Active now</span>
                  </>
                )}
              </div>
            </div>
          </div>
          
          <div className="flex items-center gap-2 text-slate-400">
            <button className="p-2 hover:bg-slate-50 rounded-full transition-colors"><Phone className="w-5 h-5" /></button>
            <button className="p-2 hover:bg-slate-50 rounded-full transition-colors"><Video className="w-5 h-5" /></button>
            <button className="p-2 hover:bg-slate-50 rounded-full transition-colors"><MoreVertical className="w-5 h-5" /></button>
          </div>
        </div>

        {/* Chat Messages (Scrollable) */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 bg-slate-50/30">
          
          <div className="flex justify-center">
            <span className="text-[11px] font-medium text-slate-400 bg-slate-100 px-3 py-1 rounded-full">Today</span>
          </div>

          {chatHistory.map((msg) => {
            const isPatient = msg.sender === 'patient';
            return (
              <div key={msg.id} className={`flex flex-col ${isPatient ? 'items-end' : 'items-start'}`}>
                <div className={`max-w-[70%] rounded-2xl px-5 py-3 text-sm leading-relaxed ${
                  isPatient 
                    ? 'bg-[#29a5f6] text-white rounded-br-sm' 
                    : 'bg-white border border-slate-100 text-slate-700 rounded-bl-sm shadow-sm'
                }`}>
                  {msg.text}
                </div>
                <div className="flex items-center gap-1 mt-1.5 px-1">
                  <span className="text-[11px] text-slate-400">{msg.time}</span>
                  {isPatient && <CheckCheck className="w-3.5 h-3.5 text-[#29a5f6]" />}
                </div>
              </div>
            );
          })}

          {/* Typing Indicator */}
          <div className="flex items-start">
            <div className="bg-white border border-slate-100 rounded-2xl rounded-bl-sm px-4 py-3 shadow-sm flex items-center gap-1">
              <div className="w-2 h-2 bg-slate-300 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
              <div className="w-2 h-2 bg-slate-300 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
              <div className="w-2 h-2 bg-slate-300 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
            </div>
          </div>
        </div>

        {/* Chat Input Area */}
        <div className="p-4 bg-white border-t border-slate-100">
          <div className="flex items-center gap-2 bg-slate-50 p-2 rounded-2xl border border-slate-200">
            <button className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-200/50 rounded-full transition-colors flex-shrink-0">
              <Paperclip className="w-5 h-5" />
            </button>
            <input 
              type="text" 
              placeholder="Type your message..." 
              className="flex-1 bg-transparent border-none text-sm focus:ring-0 px-2 outline-none"
            />
            <button className="p-2 text-slate-400 hover:text-sky-500 hover:bg-sky-50 rounded-full transition-colors flex-shrink-0">
              <SendHorizonal className="w-5 h-5" />
            </button>
          </div>
          <p className="text-center text-[10px] text-slate-400 font-medium mt-3 italic">
            Your consultation content is strictly confidential and protected by MindCare.
          </p>
        </div>

      </div>
    </div>
  );
}