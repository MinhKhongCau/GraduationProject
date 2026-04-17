'use client';

import React, { useState } from 'react';

export default function DashboardPage() {
  const [activeTab, setActiveTab] = useState<'profile' | 'history'>('profile');

  return (
    <div className="max-w-5xl mx-auto">
      {/* Tiêu đề trang */}
      <h1 className="text-2xl font-bold text-slate-900 mb-6">Patient Dashboard</h1>

      {/* Tabs Switcher */}
      <div className="flex border-b border-slate-200 mb-6 bg-slate-50/50 rounded-t-lg px-2 pt-2">
        <button
          onClick={() => setActiveTab('profile')}
          className={`px-6 py-2.5 text-sm font-semibold transition-all relative ${
            activeTab === 'profile'
              ? 'text-sky-500 border-b-2 border-sky-500 bg-white rounded-t-md shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          My Health Profile
        </button>
        <button
          onClick={() => setActiveTab('history')}
          className={`px-6 py-2.5 text-sm font-semibold transition-all relative ${
            activeTab === 'history'
              ? 'text-sky-500 border-b-2 border-sky-500 bg-white rounded-t-md shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          Appointment History
        </button>
      </div>

      {/* Content Area */}
      <div className="bg-white border border-slate-200 rounded-2xl p-8 shadow-sm min-h-[500px]">
        {activeTab === 'profile' ? (
          <div className="space-y-8 animate-in fade-in duration-500">
            <div>
              <h2 className="text-xl font-bold text-slate-900 mb-6">My Health Profile</h2>
              
              {/* Personal Information Section */}
              <div className="mb-8">
                <h3 className="text-sm font-bold text-slate-900 mb-4">Personal Information</h3>
                <div className="border border-slate-100 rounded-xl overflow-hidden">
                  <table className="w-full text-sm text-left">
                    <tbody className="divide-y divide-slate-100">
                      <tr className="hover:bg-slate-50/50 transition-colors">
                        <td className="px-6 py-4 text-slate-500 w-1/3">Date of Birth:</td>
                        <td className="px-6 py-4 text-slate-900 font-medium">1990-05-15</td>
                      </tr>
                      <tr className="hover:bg-slate-50/50 transition-colors">
                        <td className="px-6 py-4 text-slate-500">Gender:</td>
                        <td className="px-6 py-4 text-slate-900 font-medium">Female</td>
                      </tr>
                      <tr className="hover:bg-slate-50/50 transition-colors">
                        <td className="px-6 py-4 text-slate-500">Address:</td>
                        <td className="px-6 py-4 text-slate-900 font-medium">123 Calm Street, Serenity City, SA 54321</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Medical History Section */}
              <div>
                <h3 className="text-sm font-bold text-slate-900 mb-4">Medical History</h3>
                
                <div className="space-y-6">
                  <div>
                    <p className="text-xs font-bold text-slate-700 mb-3">Chronic Conditions:</p>
                    <div className="flex flex-wrap gap-2">
                      <span className="px-4 py-1.5 bg-sky-50 text-sky-600 rounded-full text-xs font-semibold border border-sky-100">
                        Seasonal Allergies
                      </span>
                    </div>
                  </div>

                  <div>
                    <p className="text-xs font-bold text-slate-700 mb-3">Allergies:</p>
                    <div className="flex flex-wrap gap-2">
                      <span className="px-4 py-1.5 bg-rose-50 text-rose-500 rounded-full text-xs font-semibold border border-rose-100">
                        Pollen
                      </span>
                      <span className="px-4 py-1.5 bg-rose-50 text-rose-500 rounded-full text-xs font-semibold border border-rose-100">
                        Dust
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center h-[400px] text-slate-400 animate-in fade-in duration-500">
            <div className="bg-slate-50 p-4 rounded-full mb-4">
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p className="text-sm font-medium">No appointment history found.</p>
          </div>
        )}
      </div>
    </div>
  );
}